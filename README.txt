# Insight Backend

This project contains the backend implementation for the Insight application. It provides APIs and services to support the application's functionality.

## Features

- RESTful API endpoints
- Database integration
- Authentication and authorization
- Logging and error handling

## Requirements

- Python 3.8+
- Dependencies listed in `requirements.txt`

## Setup

1. Clone the repository:
  ```
  git clone <repository-url>
  cd insight-backend
  ```

2. Create a virtual environment and activate it:
  ```
  python3 -m venv venv
  source venv/bin/activate
  ```

3. Install dependencies:
  ```
  pip install -r requirements.txt
  ```

4. Configure environment variables:
  - Copy `.env.example` to `.env` and update the values as needed.

5. Run database migrations:
  ```
  python manage.py migrate
  ```

6. Start the development server:
  ```
  python manage.py runserver
  ```