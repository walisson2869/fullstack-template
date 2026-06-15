'use client'

import { useEffect, useLayoutEffect, useRef, useState, useCallback } from 'react'

export interface WsEnvelope {
  type: string
  payload: unknown
}

export interface UseWebSocketOptions {
  url: string
  token?: string
  onMessage?: (envelope: WsEnvelope) => void
  maxRetries?: number
}

export interface UseWebSocketReturn {
  isConnected: boolean
  lastMessage: WsEnvelope | null
  send: (envelope: WsEnvelope) => void
}

export function useWebSocket({
  url,
  token,
  onMessage,
  maxRetries = 10,
}: UseWebSocketOptions): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WsEnvelope | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const retriesRef = useRef(0)
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const unmountedRef = useRef(false)

  // Refs holding the latest option values and connect closure.
  // Updated in useLayoutEffect (after render, before effects) — never during render.
  const onMessageRef = useRef(onMessage)
  const maxRetriesRef = useRef(maxRetries)
  const connectRef = useRef<() => void>(() => {})

  useLayoutEffect(() => {
    onMessageRef.current = onMessage
    maxRetriesRef.current = maxRetries
  })

  useLayoutEffect(() => {
    connectRef.current = () => {
      if (unmountedRef.current) return

      const wsUrl = token ? `${url}?token=${encodeURIComponent(token)}` : url
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        if (unmountedRef.current) return
        setIsConnected(true)
        retriesRef.current = 0
      }

      ws.onmessage = (event: MessageEvent) => {
        if (unmountedRef.current) return
        try {
          const envelope = JSON.parse(event.data as string) as WsEnvelope
          setLastMessage(envelope)
          onMessageRef.current?.(envelope)
        } catch {
          // Ignore non-JSON frames.
        }
      }

      ws.onclose = () => {
        if (unmountedRef.current) return
        setIsConnected(false)
        if (retriesRef.current < maxRetriesRef.current) {
          const delay = Math.min(1000 * 2 ** retriesRef.current, 30_000)
          retriesRef.current++
          retryTimerRef.current = setTimeout(() => connectRef.current(), delay)
        }
      }

      ws.onerror = () => ws.close()
    }
  })

  useEffect(() => {
    unmountedRef.current = false
    connectRef.current()

    return () => {
      unmountedRef.current = true
      if (retryTimerRef.current !== null) {
        clearTimeout(retryTimerRef.current)
        retryTimerRef.current = null
      }
      wsRef.current?.close()
    }
  }, [])

  const send = useCallback((envelope: WsEnvelope) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(envelope))
    }
  }, [])

  return { isConnected, lastMessage, send }
}
