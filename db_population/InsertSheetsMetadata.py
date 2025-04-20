from googleapiclient.discovery import build
from google.oauth2 import service_account
import pandas as pd
from utils.db_manager import add_or_update_document


# Converts df dict into dict format insert_or_update_document is prepared for
#TODO Data should follow this format, this method converts our df dict data into a dict that matches this
#data = {
#    "file_name": "test_file.pdf",
#    "title": "Test Document",
#    "abstract": "A simple test case",
#    "category": "Test",
#   "publish_date": "2025-01-01",
#    "source": "Unit Test",
#    "region": "Global",
#    "s3_file": "s3://test-bucket/test_file.pdf",
#    "s3_file_preview": "s3://test-bucket/test_file_preview.webm",
#    "pdf_link": "https://example.com/test_file.pdf",
#    "authors": ["John Doe"],
#    "keywords": ["Machine Learning", "AI"],
#}
def dict_to_dict(dict):
    date_published = format_datetime(dict["Year"], dict["Month"], dict["Day"])
    new_dict = {
        "file_name": dict["Filename"],
        "title": dict["Title"],
        "abstract": dict["Abstract"],
        "category": dict["Resource Type"].split(","),
        "publish_date": date_published,
        "source": dict["Publication"],
        "region": dict["Region"].split(","),
        "pdf_link": dict["Link"],
        "authors": dict["Authors"].split(','),
        "keywords": dict["Subject Keywords"].split(','),
    }

    return new_dict


def format_datetime(year, month, day):
    try:
        # Convert to int, will raise ValueError if invalid
        y = int(year)
        m = int(month)
        d = int(day)
        return f"{y:04d}-{m:02d}-{d:02d}T00:00:00"
    except (TypeError, ValueError):
        return None  # or a safe default like "1900-01-01T00:00:00"


def main():
    # Authenticate
    SERVICE_ACCOUNT_FILE = "env.json"
    SCOPES = ["https://www.googleapis.com/auth/spreadsheets.readonly"]


    creds = service_account.Credentials.from_service_account_file(SERVICE_ACCOUNT_FILE, scopes=SCOPES)


    SHEET_ID = "1dAVDBNL23_ew6yJ-Cd8ACuMkLbrOXhaRFTvGgaRm0tI"
    #Would be great if this could be dynamic somehow, will look into
    RANGE = "Files!A1:M488"

    service = build("sheets", "v4", credentials=creds)
    sheet = service.spreadsheets()
    result = sheet.values().get(spreadsheetId=SHEET_ID, range=RANGE).execute()
    data = result.get("values", [])

    # Convert to DataFrame
    df = pd.DataFrame(data[1:], columns=data[0])  # Assuming first row is header
    dict_list = df.to_dict(orient="records")
    print("reaches here")
    print(list(dict_list[10].keys()))
    for row_dict in dict_list:

        #Bens insert_document method takes in a dictionary, so no ned to convert it into json file
        #as long as the dictionary has all the information with correct key names and such, we should be good
        try:
            add_or_update_document(dict_to_dict(row_dict))
        except ValueError as ve:
            print(f"Validation error: {ve}")
        except Exception as e:
            print(f"Unexpected error: {e}")


if __name__ == "__main__":
    main()