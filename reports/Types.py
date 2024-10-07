from typing import TypedDict


# Type dict for firebase data
class PortfolioRecord(TypedDict):
    date: str
    skill: str
    score: float
    ai_note: str
    preview_note: str
    reflection: str


class Student(TypedDict):
    name: str
    line_id: str
    handedness: str
    portfolio: list[PortfolioRecord]
