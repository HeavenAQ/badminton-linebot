from ReportGenerator import ReportGenerator, WorkbookHandler
from GoogleDriveHandler import SecretManager, GoogleDriveHandler

SPECIFIED_DATES = [
    "09/16",
    "09/23",
    "09/30",
    "10/07",
    "10/14",
    "10/21",
    "10/28",
    "11/04",
    "11/11",
    "11/18",
    "11/25",
    "12/02",
    "12/09",
    "12/16",
    "12/23",
    "12/30",
    "01/06",
]


def create_report():
    secret_data = get_gcp_secret("firebase-credentials")
    report_generator = ReportGenerator(
        firestore_credentials=secret_data,
        root_collection="users",
        output_path="./student_portfolio_records.xlsx",
    )
    report_generator.generate_students_records()
    report_generator.workbook = WorkbookHandler("./student_average.xlsx")
    report_generator.generate_average_and_median_score_report(SPECIFIED_DATES)


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
    gdh.delete_file("student_average.xlsx")
    gdh.upload_file("./student_portfolio_records.xlsx")
    gdh.upload_file("./student_average.xlsx")


def main():
    # create the report
    create_report()
    gcp_secret = get_gcp_secret("google-drive-credentials")
    upload_report_to_google_drive(gcp_secret)


if __name__ == "__main__":
    main()
print("Done.")
