# BensTask

In this project, I used Golang without any framework to ensure longevity and maintain the quality of the code. Although using frameworks like GIN or CHI could have eased the workload, I opted for a more bare-metal approach.

## Key Features
- **Parser Simulation:** Implemented a simulation of a parser that requests new data every X seconds. The parser is notified immediately in the event of a new upload.
- **Ease of Use:** The database instance is already set up, and I’ve included a Postman file with predefined queries for easy testing.
- **Authorization:** I intended to implement JWT authentication but decided against it due to time constraints. Instead, I used the `Authorization` header to pass the user, switching between `123` and `789` for testing purposes.
- **SQLite Database:** For simplicity, the project runs on an SQLite database. Though uploading data directly into SQL is not best practice, I followed this method to meet the task’s requirements. Ideally, I would have used a dedicated data server connected to the API.

## How to Run
1. **Build the project:**
   ```bash
   make build
    ```
2. **Build the project:**
   ```bash
   make run 
   ```
Although the setup works well for this task, in a real-world scenario, I would separate the data management from the API for better scalability and performance.