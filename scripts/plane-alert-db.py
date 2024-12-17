import pandas as pd
from dotenv import load_dotenv
from sqlalchemy import create_engine
import os


def main():
    load_dotenv()

    df = get_data()
    engine = get_db_connection()
    table_name = "interesting_aircraft"

    try:
        df.to_sql(name=table_name, con=engine, if_exists="replace", index=False)
        print(f"Data successfully written to table '{table_name}'")
    except Exception as e:
        print(f"Error writing to database: {e}")

    engine.dispose()


def get_data():
    plane_db_url = os.getenv("PLANE_DB_URL")
    image_db_url = os.getenv("IMAGE_DB_URL")

    df_planes = pd.read_csv(plane_db_url)
    df_images = pd.read_csv(image_db_url)

    df_combined = pd.merge(df_planes, df_images, on="$ICAO", how="left")

    df_renamed = df_combined.rename(columns=get_column_mapping())

    return df_renamed


def get_db_connection():
    DB_NAME = os.getenv("DB_NAME")
    DB_USER = os.getenv("DB_USER")
    DB_PASSWORD = os.getenv("DB_PASSWORD")
    DB_HOST = os.getenv("DB_HOST")
    DB_PORT = os.getenv("DB_PORT")

    DATABASE_URL = (
        f"postgresql+psycopg2://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
    )

    engine = create_engine(DATABASE_URL)

    return engine


def get_column_mapping():
    column_mapping = {
        "$ICAO": "icao",
        "$Registration": "registration",
        "$Operator": "operator",
        "$Type": "type",
        "$ICAO Type": "icao_type",
        "#CMPG": "group",
        "$Tag 1": "tag1",
        "$#Tag 2": "tag2",
        "$#Tag 3": "tag3",
        "Category": "category",
        "$#Link": "link",
        "#ImageLink": "image_link_1",
        "#ImageLink2": "image_link_2",
        "#ImageLink3": "image_link_3",
        "#ImageLink4": "image_link_4",
    }

    return column_mapping


if __name__ == "__main__":
    main()
