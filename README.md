# Word Image Uploader

A modern web application built with Go that helps organize and upload word-related images to AWS S3 and store their metadata in MongoDB.

## Features

- Upload folders containing word images organized by letters (A/, B/, etc.)
- Automatically uploads images to AWS S3 with proper folder structure
- Stores metadata in MongoDB for easy retrieval
- Modern, responsive web interface
- Real-time upload progress tracking
- Built with Go backend and vanilla JavaScript frontend

## Prerequisites

- Go 1.19 or later
- MongoDB Atlas account
- AWS S3 bucket
- Node.js and npm (optional, for development)

## Configuration

Create a `.env` file in the root directory with the following variables:

```env
MONGODB_URI=your_mongodb_connection_string
AWS_REGION=your_aws_region
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_BUCKET_NAME=your_bucket_name
```

## Directory Structure

Your image folders should be organized like this:
```
images/
  A/
    airplane.png
    apple.png
  B/
    banana.png
    book.png
  ...
```

## Installation

1. Clone the repository:
git clone <repository-url>
cd goupload
```

2. Install Go dependencies:
go mod tidy
```

3. Run the application:
go run main.go
```

4. Open your browser and navigate to:
http://localhost:8080
```

## Usage

1. Click "Choose Folder" to select your organized image directory
2. The application will automatically detect the letter-based structure
3. Click "Upload Images" to start the upload process
4. View the results with links to uploaded images

## Technologies Used

- Backend:
  - Go
  - AWS SDK v2
  - MongoDB Go Driver
  - Godotenv

- Frontend:
  - HTML5
  - CSS3
  - Vanilla JavaScript
  - Font Awesome icons

## License

MIT License

## Author

Duypresso. Built with ❤️ using Go
