import { renderHook, act } from '@testing-library/react'
import { beforeEach, afterEach, describe, it, expect, vi } from 'vitest'
import { useWebSocket } from './useWebSocket'

// Minimal WebSocket mock that captures handlers and lets tests trigger them.
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.OPEN
  url: string
  onopen: ((e: Event) => void) | null = null
  onmessage: ((e: MessageEvent) => void) | null = null
  onclose: ((e: CloseEvent) => void) | null = null
  onerror: ((e: Event) => void) | null = null

  static instances: MockWebSocket[] = []

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  send = vi.fn()
  close = vi.fn(() => {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.(new CloseEvent('close'))
  })

  // Test helpers
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN
    this.onopen?.(new Event('open'))
  }
  simulateMessage(data: string) {
    this.onmessage?.(new MessageEvent('message', { data }))
  }
  simulateClose() {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.(new CloseEvent('close'))
  }
}

beforeEach(() => {
  MockWebSocket.instances = []
  vi.useFakeTimers()
  vi.stubGlobal('WebSocket', MockWebSocket)
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('useWebSocket', () => {
  it('connects to the correct URL without a token', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    expect(MockWebSocket.instances).toHaveLength(1)
    expect(MockWebSocket.instances[0].url).toBe('ws://localhost:8080/ws')
  })

  it('appends token as query param when provided', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws', token: 'abc123' }))
    expect(MockWebSocket.instances[0].url).toBe('ws://localhost:8080/ws?token=abc123')
  })

  it('sets isConnected true after open', () => {
    const { result } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    act(() => MockWebSocket.instances[0].simulateOpen())
    expect(result.current.isConnected).toBe(true)
  })

  it('sets isConnected false after close', () => {
    const { result } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    act(() => MockWebSocket.instances[0].simulateOpen())
    act(() => MockWebSocket.instances[0].simulateClose())
    expect(result.current.isConnected).toBe(false)
  })

  it('calls onMessage and updates lastMessage when a message arrives', () => {
    const onMessage = vi.fn()
    const { result } = renderHook(() =>
      useWebSocket({ url: 'ws://localhost:8080/ws', onMessage }),
    )
    const data = JSON.stringify({ type: 'ping', payload: null })
    act(() => MockWebSocket.instances[0].simulateMessage(data))

    expect(onMessage).toHaveBeenCalledWith({ type: 'ping', payload: null })
    expect(result.current.lastMessage).toEqual({ type: 'ping', payload: null })
  })

  it('ignores non-JSON messages without throwing', () => {
    const onMessage = vi.fn()
    renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws', onMessage }))
    expect(() =>
      act(() => MockWebSocket.instances[0].simulateMessage('not json')),
    ).not.toThrow()
    expect(onMessage).not.toHaveBeenCalled()
  })

  it('reconnects after close with exponential backoff', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    expect(MockWebSocket.instances).toHaveLength(1)

    act(() => MockWebSocket.instances[0].simulateClose())
    // First retry: 1000ms * 2^0 = 1000ms
    act(() => vi.advanceTimersByTime(1000))
    expect(MockWebSocket.instances).toHaveLength(2)
  })

  it('does not reconnect beyond maxRetries', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws', maxRetries: 1 }))

    // First close triggers retry 0 → opens connection 2
    act(() => MockWebSocket.instances[0].simulateClose())
    act(() => vi.advanceTimersByTime(1000))
    expect(MockWebSocket.instances).toHaveLength(2)

    // Second close — retry count reached, no more reconnects
    act(() => MockWebSocket.instances[1].simulateClose())
    act(() => vi.advanceTimersByTime(30000))
    expect(MockWebSocket.instances).toHaveLength(2)
  })

  it('closes the WebSocket and clears timers on unmount', () => {
    const { unmount } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    const ws = MockWebSocket.instances[0]

    // Trigger a pending retry before unmounting
    act(() => ws.simulateClose())
    unmount()

    // No new connection should appear after timers fire
    act(() => vi.advanceTimersByTime(30000))
    expect(MockWebSocket.instances).toHaveLength(1)
    expect(ws.close).toHaveBeenCalled()
  })

  it('send() sends JSON when the socket is open', () => {
    const { result } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080/ws' }))
    const ws = MockWebSocket.instances[0]
    act(() => ws.simulateOpen())

    act(() => result.current.send({ type: 'ping', payload: null }))
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: 'ping', payload: null }))
  })
})
