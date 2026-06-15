package com.company.template.websocket

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

/**
 * Typed wire format for all WebSocket messages, mirroring the backend Envelope struct.
 * [payload] is a raw [JsonElement] because message types vary; callers decode based on [type].
 */
@Serializable
data class WsEnvelope(
    val type: String,
    val payload: JsonElement? = null,
)
