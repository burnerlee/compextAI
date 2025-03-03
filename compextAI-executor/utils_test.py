from utils import create_pydantic_model_from_dict

schema = {
  "type": "json_schema",
  "json_schema": {
    "name": "CodeChanges",
    "schema": {
      "type": "object",
      "required": [
        "changed_files",
        "additional_message"
      ],
      "properties": {
        "changed_files": {
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "file_path",
              "content"
            ],
            "properties": {
              "content": {
                "type": "string",
                "description": "The new content of the file with the code changes"
              },
              "file_path": {
                "type": "string",
                "description": "The path to the file that has changed"
              }
            },
            "additionalProperties": False
          },
          "description": "The list of files that have changed with their new content"
        },
        "additional_message": {
          "type": "string",
          "description": "An additional message to the user as a response to their message"
        }
      },
      "additionalProperties": False
    },
    "strict": True
  }
}

model = create_pydantic_model_from_dict(schema["json_schema"]["name"], schema["json_schema"]["schema"])
# print all the fields
# print(model.model_fields)
