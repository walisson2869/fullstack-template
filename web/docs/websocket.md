---
topic: websocket
last_verified: 2026-06-15
sources:
  - lib/useWebSocket.ts
  - lib/useWebSocket.test.ts
---

# WebSocket

## Overview

The `useWebSocket` hook in `lib/useWebSocket.ts` opens a WebSocket connection to the
backend `GET /ws` endpoint, passes the Firebase ID token as a query parameter, and
reconnects automatically with exponential backoff when the connection drops.

## Typed message envelope

```ts
export interface WsEnvelope {
  type: string      // dot-separated event name, e.g. "job.completed"
  payload: unknown  // shape determined by type; cast on the receiving end
}
```

## useWebSocket hook

```ts
import { useWebSocket } from '@/lib/useWebSocket'

const { isConnected, lastMessage, send } = useWebSocket({
  url: 'ws://localhost:8080/ws',
  token: firebaseIdToken,       // optional — required by backend in staging/prod
  onMessage: (envelope) => {    // optional callback for each message
    if (envelope.type === 'job.completed') { /* ... */ }
  },
  maxRetries: 10,               // optional, default 10
})
```

### Options

| Option | Type | Default | Description |
|---|---|---|---|
| `url` | `string` | required | WebSocket URL (no token suffix needed) |
| `token` | `string` | `undefined` | Firebase ID token — appended as `?token=` |
| `onMessage` | `(e: WsEnvelope) => void` | `undefined` | Called for each valid message |
| `maxRetries` | `number` | `10` | Maximum reconnection attempts |

### Return value

| Field | Type | Description |
|---|---|---|
| `isConnected` | `boolean` | `true` while the socket is open |
| `lastMessage` | `WsEnvelope \| null` | Most recently received envelope |
| `send` | `(e: WsEnvelope) => void` | Send a message; no-op when disconnected |

## Reconnection

After a disconnect, the hook retries with exponential backoff capped at 30 seconds:

```text
delay = min(1000ms × 2^retryCount, 30 000ms)
```

Retry count resets to 0 on a successful reconnect. No further retries are attempted
after `maxRetries` failures. Pending retry timers are cleared on unmount.

## Authentication

Append the Firebase ID token as a query parameter:

```ts
useWebSocket({ url: 'ws://localhost:8080/ws', token: idToken })
// → connects to ws://localhost:8080/ws?token=<encoded-token>
```

The backend (`GET /ws`) rejects connections without a valid token with HTTP 401 before
the WebSocket upgrade completes.

## `"use client"` requirement

`useWebSocket` is a Client Component hook — it requires `"use client"` and must not
be imported from Server Components. Wrap it in a Client Component that receives
the token as a prop from a Server Component.

```tsx
// app/live/LiveFeed.tsx  (Server Component — fetches token server-side)
import LiveFeedClient from './LiveFeedClient'
export default async function LiveFeed() {
  const token = await getFirebaseToken()
  return <LiveFeedClient token={token} />
}

// app/live/LiveFeedClient.tsx
'use client'
import { useWebSocket } from '@/lib/useWebSocket'
export default function LiveFeedClient({ token }: { token: string }) {
  const { isConnected, lastMessage } = useWebSocket({ url: process.env.NEXT_PUBLIC_WS_URL!, token })
  // ...
}
```

## Testing

Tests live in `lib/useWebSocket.test.ts` and use Vitest + `@testing-library/react`
with a `MockWebSocket` class injected via `vi.stubGlobal('WebSocket', MockWebSocket)`.
Fake timers (`vi.useFakeTimers`) drive reconnection delays synchronously.

Covered cases:
- Correct URL construction (with and without token)
- `isConnected` state transitions on open/close
- `onMessage` callback and `lastMessage` state update
- Non-JSON message tolerance
- Exponential backoff reconnection
- `maxRetries` limit
- Cleanup on unmount (close + timer clear)
- `send()` delegates to `WebSocket.send()`
