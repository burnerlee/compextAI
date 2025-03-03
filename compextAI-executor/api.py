import fastapi
import uvicorn
import os
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import openai_models as openai
import anthropic_models as anthropic
import json
import litellm_base as litellm
app = fastapi.FastAPI()


@app.get("/")
def read_root():
    return {"pong"}

class ChatCompletionRequest(BaseModel):
    """
    Request body for the chat completion endpoint.
    """
    api_keys: dict
    model: str
    messages: list[dict]
    temperature: float = 0.5
    timeout: int = 600
    max_tokens: int = 10000
    max_completion_tokens: int = 10000
    response_format: dict = None
    system_prompt: str = None
    tools: list[dict] = None

@app.post("/chatcompletion/openai")
def chat_completion_openai(request: ChatCompletionRequest):
    try:
        response = openai.chat_completion(request.api_keys, request.model, request.messages, request.temperature, request.timeout, request.max_completion_tokens, request.response_format, request.tools)
        return JSONResponse(status_code=200, content=json.loads(response))
    except Exception as e:
        print(e)
        return JSONResponse(status_code=500, content={"error": str(e)})
    
@app.post("/chatcompletion/anthropic")
def chat_completion_anthropic(request: ChatCompletionRequest):
    try:
        response = anthropic.chat_completion(request.api_keys, request.system_prompt, request.model, request.messages, request.temperature, request.timeout, request.max_tokens, request.response_format, request.tools)
        return JSONResponse(status_code=200, content=json.loads(response))
    except Exception as e:
        print(e)
        return JSONResponse(status_code=500, content={"error": str(e)})
    
@app.post("/chatcompletion/litellm")
def chat_completion_litellm(request: ChatCompletionRequest):
    try:
        response = litellm.chat_completion(request.api_keys, request.model, request.messages, request.temperature, request.timeout, request.max_completion_tokens, request.response_format, request.tools)
        return JSONResponse(status_code=200, content=json.loads(response))
    except Exception as e:
        print(e)
        return JSONResponse(status_code=500, content={"error": str(e)})

if __name__ == "__main__":
    port = 8889
    if os.getenv("SERVER_PORT"):
        port = int(os.getenv("SERVER_PORT"))
    uvicorn.run(app, host="0.0.0.0", port=port)

