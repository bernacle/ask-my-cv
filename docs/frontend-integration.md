# Frontend Integration

The existing Next.js portfolio should treat Ask My CV as a separate API service.

## Environment

```bash
NEXT_PUBLIC_ASK_MY_CV_API_URL=https://your-railway-api.up.railway.app
```

## Request

```ts
const response = await fetch(`${process.env.NEXT_PUBLIC_ASK_MY_CV_API_URL}/api/v1/ask`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    question,
    conversation_id: conversationId ?? undefined,
  }),
});
```

## UI Behavior

- Keep the chat widget in the portfolio and call this backend from the browser.
- Show a loading state while the request is pending.
- Store `conversation_id` returned by the first answer and reuse it for follow-up questions.
- Render `citations` below each answer with `source_title`, `source_path`, and `chunk_index`.
- If the API returns `429`, show a friendly "try again later" state.
- If `citations` is empty, render the answer as an unsupported/no-source response.

## CORS

Set `CORS_ALLOWED_ORIGINS` on the backend to:

```bash
https://brunosimplicio.com,http://localhost:3000,http://localhost:3001
```
