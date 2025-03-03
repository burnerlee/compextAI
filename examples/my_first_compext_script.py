import compextAI.api.api as compextAPI
import compextAI.execution as compextExecution
import compextAI.threads as compextThreads
import compextAI.params as compextParams
import compextAI.messages as compextMessages
import time
import os

API_KEY = ""

if os.environ.get("API_KEY"):
    API_KEY = os.environ.get("API_KEY")

client = compextAPI.APIClient(
    base_url="https://api.compextai.dev",
    api_key=API_KEY
)

thread: compextThreads.Thread = compextThreads.create(
    client=client,
    project_name="my-first-project",
    title="My first thread"
)

compextMessages.create(
    client=client,
    thread_id=thread.thread_id,
    messages=[
        compextMessages.Message(
            role="user",
            content="Say hello!"
        )
    ]
)

params = compextParams.retrieve(
    client=client,
    name="my-first-config",
    environment="test",
    project_name="my-first-project"
)

thread_execution = thread.execute(
    client=client,
    thread_exec_param_id=params.thread_execution_param_id,
    append_assistant_response=True
)

def wait_for_completion(thread_execution_id: str):
    status: compextExecution.ThreadExecutionStatus = compextExecution.get_thread_execution_status(
        client=client,
        thread_execution_id=thread_execution_id
    )
    if status.status == "completed":
        return
    if status.status == "failed":
        raise Exception("Thread execution failed")
    time.sleep(1)
    wait_for_completion(thread_execution_id)

wait_for_completion(thread_execution.thread_execution_id)

thread_execution_result: dict = compextExecution.get_thread_execution_response(
    client=client,
    thread_execution_id=thread_execution.thread_execution_id
)

# print(thread_execution_result['content'])
