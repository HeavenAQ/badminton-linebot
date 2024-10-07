from ReportGenerator import ReportGenerator
from GoogleDriveHandler import SecretManager, GoogleDriveHandler


def create_report():
    secret_data = get_gcp_secret("firebase-credentials")
    report_generator = ReportGenerator(
        firestore_credentials=secret_data,
        root_collection="users",
        output_path="./student_portfolio_records.xlsx",
    )
    report_generator.generate()


def get_gcp_secret(credentials: str):
    secret_manager = SecretManager()
    secret_name = secret_manager.get_secret_name_string(credentials)
    secret_data = secret_manager.access_secret_version(secret_name)
    secret_manager.close()
    return secret_data


def upload_report_to_google_drive(gcp_secret: bytes):
    SCOPES = ["https://www.googleapis.com/auth/drive"]
    gdh = GoogleDriveHandler(
        SCOPES,
        gcp_secret,
        "1gHlA1HH1lNbDFcHVf4nzvQcYFJNhqVds",
    )
    gdh.delete_file("student_portfolio_records.xlsx")
    gdh.upload_file("./student_portfolio_records.xlsx")


def main():
    # create the report
    create_report()
    gcp_secret = get_gcp_secret("google-drive-credentials")
    upload_report_to_google_drive(gcp_secret)


if __name__ == "__main__":
    main()
print("Done.")
