# File Analyzer Application

## Overview
The File Analyzer Application is a client-server application developed in Go that allows users to analyze text files. The server processes incoming text files from clients using an HTTP API, counts the number of words, characters, and lines, and returns the analysis results back to the clients.

## Project Structure
```
file-analyzer-app
├── cmd
│   ├── client
│   │   └── main.go        # Entry point for the client application
│   └── server
│       └── main.go        # Entry point for the server application
├── internal
│   ├── analyzer
│   │   └── analyzer.go    # Logic for analyzing text files
│   ├── models
│   │   └── fileanalysis.go # Structure for file analysis results
│   └── utils
│       └── fileutils.go    # Utility functions for file handling
├── pkg
│   └── network
│       └── protocol.go     # HTTP protocol definitions
├── go.mod                  # Go module definition
├── go.sum                  # Module dependency checksums
└── README.md               # Project documentation
```

## Setup Instructions
1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd file-analyzer-app
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Run the server:**
   ```
   go run cmd/server/main.go
   ```

4. **Run the client:**
   ```
   go run cmd/client/main.go <path-to-text-file1> <path-to-text-file2> ...
   ```

## Usage
- The client connects to the server and uploads one or more text files for analysis.
- The server processes the files and returns the analysis results, including the number of words, characters, and lines for each file.
- You can also use Postman to interact with the API:
  - Single file: Send a POST request to `http://localhost:8080/analyze` with a file attached in the form-data with key name "file"
  - Multiple files: Send a POST request to `http://localhost:8080/analyze-batch` with files attached in the form-data with key name "files"
  - The server will respond with a JSON containing the analysis results

## API Endpoints
- `POST /analyze` - Upload a single file for analysis
- `POST /analyze-batch` - Upload multiple files for batch analysis
- `GET /status` - Check server status

## Example
After running the client with multiple text files, you might see output like this:
```
Analysis results:

Analysis for file1.txt:
  Words: 120
  Characters: 800
  Lines: 15

Analysis for file2.txt:
  Words: 95
  Characters: 620
  Lines: 10
```

## Using Postman
1. Open Postman and create a new POST request
2. Set the URL to `http://localhost:8080/analyze-batch`
3. Go to the Body tab, select form-data
4. Add a key named "files" and select "File" from the dropdown
5. Click "Select Files" and choose multiple text files
6. Click "Send" to get the analysis results for all files

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for details.