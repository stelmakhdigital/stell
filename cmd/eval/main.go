package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/budaev/agent/internal/agent"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/guardrails"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/runtimeclient"
	"github.com/budaev/agent/internal/tools"
	"github.com/budaev/agent/internal/tools/builtin"
	"github.com/budaev/agent/pkg/config"
	"github.com/budaev/agent/pkg/eval"
	"github.com/budaev/agent/runtime/executor"
	"github.com/budaev/agent/runtime/sandbox"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "check-regression" {
		if err := checkRegression(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := runEval(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runEval(args []string) error {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	golden := fs.String("golden-set", "./eval/golden", "golden set directory")
	output := fs.String("output", "./eval/results", "output directory")
	cfgPath := fs.String("config", "configs/agent.yaml", "agent config")
	thPath := fs.String("thresholds", "configs/thresholds.yaml", "thresholds config")
	model := fs.String("model", "", "model override")
	providerName := fs.String("provider", "", "provider override")
	useJudge := fs.Bool("judge", false, "enable LLM-as-Judge")
	fixedAnswer := fs.String("fixed-answer", "", "use fixed answer instead of agent (smoke)")
	limit := fs.Int("limit", 0, "limit number of cases (0=all)")
	threshold := fs.Float64("threshold", -1, "override aggregate threshold (-1=from config)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	th, err := loadThresholds(*thPath)
	if err != nil {
		return err
	}
	if *threshold >= 0 {
		th.Aggregate = *threshold
	}

	cases, err := eval.LoadCases(*golden)
	if err != nil {
		return err
	}
	if *limit > 0 && *limit < len(cases) {
		cases = cases[:*limit]
	}

	var runner eval.Runner
	var provider llm.Provider
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}
	if *model != "" {
		cfg.LLM.Model = *model
	}
	if *providerName != "" {
		cfg.LLM.Provider = *providerName
	}

	if *fixedAnswer != "" {
		runner = eval.FixedRunner{Answer: *fixedAnswer}
	} else {
		provider, err = newProvider(cfg)
		if err != nil {
			return err
		}
		ws, _ := os.Getwd()
		client := runtimeclient.NewLocal(executor.New(sandbox.NewDocker(sandbox.DefaultPolicy())))
		reg := tools.NewRegistry()
		builtin.RegisterRuntimeTools(reg, client, ws)
		a := agent.New(
			agent.WithEventBus(eventbus.New()),
			agent.WithRegistry(reg),
			agent.WithProvider(provider),
			agent.WithModel(cfg.LLM.Model),
			agent.WithMaxLoopDepth(cfg.Agent.MaxLoopDepth),
			agent.WithTemperature(cfg.Agent.Temperature),
			agent.WithMaxTokens(cfg.Agent.MaxTokens),
			agent.WithGuardrails(guardrails.New()),
		)
		runner = eval.AgentRunner{Agent: a}
	}

	h := &eval.Harness{
		Runner:     runner,
		Thresholds: th,
		Model:      cfg.LLM.Model,
	}
	if *useJudge {
		if provider == nil {
			provider, err = newProvider(cfg)
			if err != nil {
				return err
			}
		}
		judgeModel := cfg.LLM.Model
		h.Judge = func(ctx context.Context, c eval.Case, answer string) (map[string]float64, error) {
			return eval.JudgeScore(ctx, provider, judgeModel, c, answer)
		}
	}

	rep, err := h.Run(context.Background(), cases)
	if err != nil {
		return err
	}

	outDir := filepath.Join(*output, time.Now().UTC().Format("20060102_150405"))
	if err := eval.WriteReport(outDir, rep); err != nil {
		return err
	}
	if err := eval.WriteReport(*output, rep); err != nil {
		return err
	}

	fmt.Printf("aggregate=%.4f threshold=%.4f passed=%v cases=%d out=%s\n",
		rep.Aggregate, rep.Threshold, rep.Passed, len(rep.Results), outDir)
	if !rep.Passed {
		os.Exit(1)
	}
	return nil
}

func checkRegression(args []string) error {
	fs := flag.NewFlagSet("check-regression", flag.ContinueOnError)
	results := fs.String("results", "./eval/results/results.json", "current results")
	baseline := fs.String("baseline", "./eval/results/baseline.json", "baseline results")
	tolerance := fs.Float64("tolerance", 0.02, "allowed drop")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := os.ReadFile(*results)
	if err != nil {
		return err
	}
	var current eval.Report
	if err := json.Unmarshal(data, &current); err != nil {
		return err
	}
	return eval.CheckRegression(*baseline, current, *tolerance)
}

func loadThresholds(path string) (eval.Thresholds, error) {
	th := eval.Thresholds{Aggregate: 0.85, DeterministicWeight: 0.7, JudgeWeight: 0.3}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return th, nil
		}
		return th, err
	}
	if err := yaml.Unmarshal(data, &th); err != nil {
		return th, err
	}
	return th, nil
}

func newProvider(cfg config.Config) (llm.Provider, error) {
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("AGENT_LLM_API_KEY")
	}
	switch strings.ToLower(cfg.LLM.Provider) {
	case "ollama", "":
		return llm.NewOllama(cfg.LLM.BaseURL, apiKey), nil
	case "vllm":
		return llm.NewVLLM(cfg.LLM.BaseURL, apiKey), nil
	case "openai":
		base := cfg.LLM.BaseURL
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		return llm.NewOpenAICompat("openai", base, apiKey), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", cfg.LLM.Provider)
	}
}
