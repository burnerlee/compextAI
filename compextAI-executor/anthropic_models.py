from anthropic import Anthropic
from anthropic._types import NOT_GIVEN
import instructor
from utils import create_pydantic_model_from_dict
import json

def get_client(api_key):
    return Anthropic(api_key=api_key)

def get_instructor_client(api_key):
    return instructor.from_anthropic(Anthropic(api_key=api_key))

def chat_completion(api_keys:dict, system_prompt, model, messages, temperature, timeout, max_tokens, response_format, tools):
    if response_format is None or response_format == {}:
        client = get_client(api_keys["anthropic"])
        response = client.messages.create(
        model=model,
            system=system_prompt if system_prompt else NOT_GIVEN,
            messages=messages,
            temperature=temperature,
            timeout=timeout,
            max_tokens=max_tokens,
            tools=tools if tools else NOT_GIVEN
        )
        llm_response = response.model_dump_json()
    else:
        client = get_instructor_client(api_keys["anthropic"])
        response_model = create_pydantic_model_from_dict(
            response_format["json_schema"]["name"],
            response_format["json_schema"]["schema"]
        )
        answer, complete_response = client.messages.create_with_completion(
            model=model,
            system=system_prompt,
            messages=messages,
            temperature=temperature,
            timeout=timeout,
            max_tokens=max_tokens,
            response_model=response_model,
        )
        llm_response = json.loads(complete_response.model_dump_json())
        llm_response["content"][0]["text"] = answer.model_dump_json()
        llm_response = json.dumps(llm_response)

    return llm_response
