import os
from typing import Any
from google.oauth2 import service_account
from googleapiclient.discovery import MediaFileUpload, build, json

from google.cloud import secretmanager


class SecretManager:
    def __init__(self):
        self.client = secretmanager.SecretManagerServiceClient()
        self.project_id = "linebot-nstc"

    def access_secret_version(self, name: str) -> bytes:
        """Access a specific secret version."""
        response = self.client.access_secret_version(request={"name": name})
        return response.payload.data

    def get_secret_name_string(self, secret_id: str) -> str:
        """Generate the full resource name for a secret."""
        return f"projects/{self.project_id}/secrets/{secret_id}/versions/latest"

    def close(self):
        """Close the Secret Manager client."""
        self.client.transport.close()


class GoogleDriveHandler:
    def __init__(self, scope: list[str], credentials: bytes, root_folder_id: str):
        self.scopes = scope
        self.credentials = json.loads(credentials.decode("utf-8"))
        self.service = self.init_service()
        self.root_folder_id = root_folder_id

    def init_service(self) -> Any:
        cred = service_account.Credentials.from_service_account_info(
            self.credentials, scopes=self.scopes
        )
        return build("drive", "v3", credentials=cred)

    def list_folder(self, parent_folder_id=None):
        """List folders and files in Google Drive."""
        results = (
            self.service.files()
            .list(
                q=f"'{parent_folder_id}' in parents and trashed=false"
                if parent_folder_id
                else None,
                pageSize=1000,
                fields="nextPageToken, files(id, name, mimeType)",
            )
            .execute()
        )
        items = results.get("files", [])
        print(items)

    def get_file_id(self, file_name: str) -> str | None:
        """Get the file ID in Google Drive."""
        try:
            file_id = (
                self.service.files()
                .list(
                    q=f"'{self.root_folder_id}' in parents and trashed=false and name='{file_name}'"
                )
                .execute()
                .get("files")[0]
                .get("id")
            )
            return file_id
        except IndexError:
            return None

    def delete_file(self, file_name: str):
        """Delete a file in Google Drive."""
        file_id = self.get_file_id(file_name)
        if file_id:
            self.service.files().delete(fileId=file_id).execute()
            print(f"File '{file_name}' deleted.")
        else:
            print(f"File '{file_name}' not found.")

    def upload_file(self, file_path: str):
        """Upload a file to Google Drive."""
        file_name = os.path.basename(file_path)
        file_metadata = {
            "name": file_name,
            "parents": [self.root_folder_id],
        }

        # Create MediaFileUpload object
        media = MediaFileUpload(file_path, mimetype="application/octet-stream")

        # Upload the file
        self.service.files().create(
            body=file_metadata,
            media_body=media,
        ).execute()

        print(f"File '{file_name}' uploaded.")
