package com.agent

import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

/** Minimal public-API client. Token auth (not Hands HMAC). */
class Client(private val base: String, private val token: String) {
    private val http = HttpClient.newHttpClient()

    fun createSession(message: String): String {
        val req = HttpRequest.newBuilder(URI.create("$base/v1/sessions"))
            .header("Authorization", "Bearer $token")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString("""{"message":${jsonStr(message)}}"""))
            .build()
        val body = http.send(req, HttpResponse.BodyHandlers.ofString()).body()
        val key = "\"session_id\":\""
        val i = body.indexOf(key)
        if (i < 0) return body
        val rest = body.substring(i + key.length)
        return rest.substring(0, rest.indexOf('"'))
    }

    fun cancel(id: String) {
        val req = HttpRequest.newBuilder(URI.create("$base/v1/sessions/$id/cancel"))
            .header("Authorization", "Bearer $token")
            .POST(HttpRequest.BodyPublishers.noBody())
            .build()
        http.send(req, HttpResponse.BodyHandlers.discarding())
    }

    fun hitl(id: String, decision: String) {
        val req = HttpRequest.newBuilder(URI.create("$base/v1/sessions/$id/hitl"))
            .header("Authorization", "Bearer $token")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString("""{"decision":"$decision"}"""))
            .build()
        http.send(req, HttpResponse.BodyHandlers.discarding())
    }

    private fun jsonStr(s: String) = "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"") + "\""
}
