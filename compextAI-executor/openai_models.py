import openai
from openai._types import NOT_GIVEN
import instructor
from utils import create_pydantic_model_from_dict
import json

def get_client(api_key):
    return openai.OpenAI(
        api_key=api_key
    )

def get_instructor_client(api_key):
    return instructor.from_openai(openai.OpenAI(
        api_key=api_key
    ))

def chat_completion(api_keys:dict, model:str, messages:list, temperature:float, timeout:int, max_completion_tokens:int, response_format:dict, tools:list[dict]):
    if response_format is None or response_format == {}:
        client = get_client(api_keys["openai"])
        response = client.chat.completions.create(
            model=model,
            messages=messages,
            temperature=temperature,
            timeout=timeout,
            max_completion_tokens=max_completion_tokens,
            tools=tools if tools else NOT_GIVEN
        )
        llm_response = response.model_dump_json()
    else:
        client = get_instructor_client(api_keys["openai"])
        response_model = create_pydantic_model_from_dict(
            response_format["json_schema"]["name"],
            response_format["json_schema"]["schema"]
        )
        answer, complete_response = client.chat.completions.create_with_completion(
            model=model,
            messages=messages,
            temperature=temperature,
            timeout=timeout,
            max_completion_tokens=max_completion_tokens,
            response_model=response_model,
        )

        llm_response = json.loads(complete_response.model_dump_json())
        llm_response["choices"][0]["message"]["content"] = answer.model_dump_json()
        llm_response = json.dumps(llm_response)
    return llm_response
