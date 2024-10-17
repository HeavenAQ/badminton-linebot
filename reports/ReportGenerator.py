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

        # for student data
        self.students = [*map(self.extract_student_data, self.firestore.docs)]

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

    def generate_students_records(self):
        for student in self.students:
            # skip the admin and the teacher
            if student["name"] == "Heaven" or "林國欽" in student["name"]:
                continue

            # show the students who haven't completed the reflection and/or preview_note
            missing_fields = self.get_studemtn_missing_fields(student)
            if missing_fields:
                print(missing_fields)
                # send_line_warning(sheet["line_id"], sheet["name"], missing_fields

            # Convert the student portfolio records to a DataFrame
            portfolio_df = pd.DataFrame(student["portfolio"])

            # Create a new sheet for the student
            ws = self.workbook.create_student_sheet(student)
            self.workbook.add_student_records(ws, portfolio_df)
            self.workbook.mark_missing_fields(ws)

            # create a progress chart for the student
            img_stream = self.create_student_progress_chart(
                portfolio_df, ["serve", "clear"]
            )
            img = Image(img_stream)
            ws.add_image(img, "A20")

        self.workbook.save()

    def generate_average_and_median_score_report(self, specified_dates: list[str]):
        # List of specified dates in mm/dd format
        # Convert specified_dates to a set for faster lookup
        specified_dates_set = set(specified_dates)

        # Initialize an empty list to collect all records
        all_records = []

        for student in self.students:
            # Skip the admin and the teacher
            if student["name"] == "Heaven" or "林國欽" in student["name"]:
                continue

            # For each portfolio record, extract date and score
            date_max_records = {}
            for record in student["portfolio"]:
                date_str = record["date"]  # Format is "%Y-%m-%d-%H-%M"
                # Extract month and day
                try:
                    date_obj = dt.strptime(date_str, "%Y-%m-%d-%H-%M")
                except ValueError:
                    # Handle any parsing errors
                    continue

                month_day = date_obj.strftime("%m/%d")
                if month_day in specified_dates_set:
                    date_max_record = date_max_records.get(month_day, 0)
                    if date_max_record < record["score"]:
                        date_max_records[month_day] = record["score"]

            for date, score in date_max_records.items():
                all_records.append({"date": date, "score": score})

        # Now, we have all_records containing date and score for specified dates
        # Convert to pandas DataFrame
        records_df = pd.DataFrame(all_records)

        if records_df.empty:
            print("No records found for the specified dates.")
            return

        # Group by date and compute average score
        avg_scores = records_df.groupby("date")["score"].mean().reset_index()

        # Ensure dates are sorted according to specified_dates
        avg_scores["date"] = pd.Categorical(
            avg_scores["date"], categories=specified_dates, ordered=True
        )
        avg_scores = avg_scores.sort_values("date")

        # Group by date and compute median score
        median_scores = records_df.groupby("date")["score"].median().reset_index()
        median_scores["date"] = pd.Categorical(
            median_scores["date"], categories=specified_dates, ordered=True
        )
        median_scores = median_scores.sort_values("date")

        # Now, write this data into an Excel sheet
        ws = self.workbook.wb.create_sheet("Average and Median Scores")

        # Write the header
        ws.append(["Date", "Average Score", "Median Score"])

        for avg_row, median_row in zip(
            avg_scores.itertuples(index=False), median_scores.itertuples(index=False)
        ):
            ws.append([avg_row.date, avg_row.score, median_row.score])

        # Create a line plot of the average scores over time
        fig, ax = plt.subplots(figsize=(10, 6))
        ax.plot(
            avg_scores["date"],
            avg_scores["score"],
            marker="o",
            linestyle="-",
            label="Average Score",
        )
        ax.plot(
            median_scores["date"],
            median_scores["score"],
            marker="x",
            linestyle="--",
            label="Median Score",
        )
        ax.set_title("Average Scores Over Time")
        ax.set_xlabel("Date")
        ax.set_ylabel("Average Score")
        ax.set_xticks(avg_scores["date"])
        ax.set_xticklabels(avg_scores["date"], rotation=45)

        plt.tight_layout()
        img_stream = io.BytesIO()
        plt.savefig(img_stream, format="png")
        plt.close()
        img_stream.seek(0)
        img = Image(img_stream)
        ws.add_image(img, "D2")  # Adjust the cell position as needed

        # Save the workbook
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
