from datetime import datetime as dt
from typing import Any
import io
from firebase_admin import credentials, json
from firebase_admin import firestore
import firebase_admin
from google.cloud.firestore import DocumentSnapshot
from google.cloud.firestore_v1.stream_generator import StreamGenerator
from openpyxl import Workbook
import matplotlib.pyplot as plt
from openpyxl.drawing.image import Image
from openpyxl.formatting.rule import CellIsRule
from openpyxl.styles import PatternFill
import pandas as pd

from Types import PortfolioRecord, Student


class FirestoreHandler:
    def __init__(self, firebase_cred_path: bytes, root_collection: str):
        self.__firebase_cred_path = firebase_cred_path
        self.__root_collection = root_collection
        self.docs = self.init_docs()

    def init_docs(self) -> StreamGenerator[DocumentSnapshot]:
        cred_json = json.loads(self.__firebase_cred_path.decode("utf-8"))
        cred = credentials.Certificate(cred_json)
        app = firebase_admin.initialize_app(cred)
        db = firestore.client(app)
        return db.collection(self.__root_collection).stream()


class ReportGenerator:
    def __init__(
        self, firestore_credentials: bytes, root_collection: str, output_path: str
    ):
        # for firebase admin sdk
        self.firestore = FirestoreHandler(firestore_credentials, root_collection)

        # for workbook and openpyxl
        self.workbook = WorkbookHandler(output_path)

    def init_workbook(self):
        wb = Workbook()
        if wb.active:
            wb.remove(wb.active)
        return wb

    def extract_student_data(self, doc: DocumentSnapshot) -> Student:
        doc_dict = doc.to_dict()
        if not doc_dict:
            raise ValueError("Document is empty")

        student_data = Student(
            {
                "name": doc_dict["Name"],
                "line_id": doc_dict["Id"],
                "handedness": "right" if doc_dict["Handedness"] == 1 else "left",
                "portfolio": [
                    PortfolioRecord(
                        {
                            "date": record["DateTime"],
                            "skill": "serve",  # Assuming the skill is serve
                            "score": record["Rating"],
                            "ai_note": record["AINote"],
                            "preview_note": ""
                            if "尚未填寫" in (note := record["PreviewNote"])
                            else note,
                            "reflection": ""
                            if "尚未填寫" in (note := record["Reflection"])
                            else note,
                        }
                    )
                    for record in (
                        doc_dict["Portfolio"]["Serve"] | doc_dict["Portfolio"]["Clear"]
                    ).values()
                ],
            }
        )
        # sort by date to ensure the order is correct
        student_data["portfolio"].sort(
            key=lambda x: dt.strptime(
                x["date"],
                "%Y-%m-%d-%H-%M",
            )
        )
        return student_data

    def get_studemtn_missing_fields(self, sheet: Student) -> list[str]:
        # initialize the list
        student_missing_fields = []
        action_dict = {
            "reflection": "「本週學習反思」",
            "preview_note": "「課前動作檢測」",
        }

        # show the students who haven't completed the reflection and/or preview_note
        for record in sheet["portfolio"]:
            missing_fields = []
            if not record["reflection"]:
                missing_fields.append("reflection")
            if not record["preview_note"]:
                missing_fields.append("preview_note")

            if missing_fields:
                student_missing_fields.append(
                    f"{sheet['name']}：「{record['date']}」- {'及'.join(map(lambda x: action_dict[x], missing_fields))}"
                )
        return student_missing_fields

    def create_student_progress_chart(
        self, portfolio_df: pd.DataFrame, skills: list[str]
    ) -> io.BytesIO:
        # Create a figure with n subplots
        _, axs = plt.subplots(len(skills), 1, figsize=(6, 8))
        for i, skill in enumerate(skills):
            # Plot the skills
            current_df = portfolio_df[portfolio_df["skill"] == skill]
            axs[i].plot(
                current_df["date"].apply(lambda x: "/".join(x.split("-")[1:3])),
                current_df["score"],
                marker="o",
                color="b",
                label=f"{skill.capitalize()} Score",
            )

            # Set the title, x-axis label, and y-axis label
            axs[i].set_title(f"{skill.capitalize()} Skill Progress")
            axs[i].set_xlabel("Date")
            axs[i].set_ylabel("Score")

        plt.tight_layout()  # for better spacing
        img_stream = io.BytesIO()
        plt.savefig(img_stream, format="png")
        plt.close()
        img_stream.seek(0)
        return img_stream

    def generate(self):
        for doc in self.firestore.docs:
            sheet = self.extract_student_data(doc)

            # skip the admin and the teacher
            if sheet["name"] == "Heaven" or "林國欽" in sheet["name"]:
                continue

            # show the students who haven't completed the reflection and/or preview_note
            missing_fields = self.get_studemtn_missing_fields(sheet)
            if missing_fields:
                print(missing_fields)
                # send_line_warning(sheet["line_id"], sheet["name"], missing_fields

            # Convert the student portfolio records to a DataFrame
            portfolio_df = pd.DataFrame(sheet["portfolio"])

            # Create a new sheet for the student
            ws = self.workbook.create_student_sheet(sheet)
            self.workbook.add_student_records(ws, portfolio_df)
            self.workbook.mark_missing_fields(ws)

            # create a progress chart for the student
            img_stream = self.create_student_progress_chart(
                portfolio_df, ["serve", "clear"]
            )
            img = Image(img_stream)
            ws.add_image(img, "A20")

        self.workbook.save()


class WorkbookHandler:
    def __init__(self, output_path: str):
        self.wb = self.init_workbook()
        self.output_path = output_path

    def init_workbook(self):
        wb = Workbook()
        if wb.active:
            wb.remove(wb.active)
        return wb

    def create_student_sheet(self, sheet: Student) -> Any:
        ws = self.wb.create_sheet(sheet["name"])
        ws.append(["Name", sheet["name"]])
        ws.append(["Line ID", sheet["line_id"]])
        ws.append(["Handedness", sheet["handedness"]])
        ws.append([])  # Empty row
        ws.append(["Date", "Skill", "Score", "AI Note", "Preview Note", "Reflection"])
        return ws

    def add_student_records(self, ws: Any, portfolio_df: pd.DataFrame):
        for row in portfolio_df.itertuples(index=False):
            ws.append(row)

    def mark_missing_fields(self, ws: Any):
        red_fill = PatternFill(
            start_color="FFC7CE", end_color="FFC7CE", fill_type="solid"
        )
        # Apply the formatting rule for blanks in columns A to E, starting from row 5
        ws.conditional_formatting.add(
            f"A5:E{ws.max_row}",
            CellIsRule(operator="equal", formula=['""'], fill=red_fill),
        )

    def save(self):
        self.wb.save(self.output_path)
