
# Test Data Generation Program

The Test Data Generation Program is a Golang-based application designed to facilitate the generation of test data for various scenarios. It creates a simulated video stream with timestamps and captures still images at defined intervals. The captured images are securely uploaded to a FileZilla server using the FTPS (File Transfer Protocol Secure) protocol. This program is ideal for testing edge server setups or any application that requires realistic and customizable test data.

## Table of Contents
- [Features](#features)
- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Configuration](#configuration)
    - [Usage](#usage)
    - [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)
- [Support](#support)
- [Disclaimer](#disclaimer)
- [Acknowledgments](#acknowledgments)

## Features

- **Test Video Generation**: Generate a video stream with customizable parameters such as resolution, frame rate, and duration.
- **Timestamps**: Add timestamps to the video stream, providing a clear reference point for analyzing the test data.
- **Still Image Capture**: Capture still images at specified intervals from the generated video stream.
- **Secure FTPS Upload**: Utilize FTPS protocol to securely transfer the captured images to a FileZilla server.
- **Configuration Flexibility**: Customize the program behavior by providing a configuration file with various options.
- **Error Handling and Logging**: Comprehensive error handling and logging mechanisms ensure easy debugging and troubleshooting.
- **Docker Containerization**: The program can be easily containerized using Docker for seamless deployment and scalability.
-


## Program Components documentation
#### `establishFTPSConnection` Function

The `establishFTPSConnection` function is a key part of this program. It's designed to establish an FTPS (FTP Secure) connection with a remote server using the lftp utility.

### establishFTPSConnection Function Signature
```go
func establishFTPSConnection(config Config) error
```
### establishFTPSConnection Parameters
This function accepts a single parameter:
- `config`: An instance of the `Config` struct, containing the FTPS configuration details, including the FTPS server address, port, user credentials, and destination upload directory.

### establishFTPSConnection Function Behavior
The function first attempts to establish an FTPS connection using the information provided in the `config` parameter. If it fails, it waits for a predefined interval (`retryInterval`) and then tries again. This process is repeated for a maximum number of times specified by `maxRetries`.

If the function manages to successfully establish the FTPS connection, it immediately returns `nil`, indicating no error occurred. If the function exhausts all retry attempts without successfully establishing a connection, it returns an error stating that it failed to establish an FTPS connection after a certain number of attempts.

### establishFTPSConnection function Error Handling
If the function cannot establish the FTPS connection after all retry attempts, it does not halt the execution of the entire program. Instead, it logs an error message and returns an error object. The main function can then decide how to handle this situation, e.g., it could ignore the error and proceed with the program, or it could decide to stop execution altogether.

### establishFTPSConnection function Usage
This function should be called when you're ready to establish an FTPS connection, typically after you've prepared the data (e.g., video and snapshots) you want to upload. Because the function includes a retry mechanism, you don't need to manually retry the connection attempt in case of temporary network issues or server downtime.

### Example
```go
err := establishFTPSConnection(config)
if err != nil {
    log.Fatalf("Failed to establish FTPS connection: %v", err)
}
```
In this example, if the `establishFTPSConnection` function returns an error, the error is logged and the program terminates.


##########################################

## Getting Started

These instructions will guide you through the installation, configuration, and usage of the Test Data Generation Program.

### Prerequisites

- Go programming language (v1.16 or later) must be installed on your system.
- Access to a FileZilla server with FTPS enabled for secure image storage.

### Installation

1. Clone the repository to your local machine:
   ```
   git clone [repository_url]
   ```
2. Navigate to the project directory:
   ```
   cd Test-Data-Generation-Program
   ```
3. Install the required dependencies using Go modules:
   ```
   go mod download
   ```

### Configuration

1. Create a configuration file named `configuration.json` in the project directory.
2. Open the configuration file in a text editor and provide the following information:

  ```json
{
  "resolution": "3840x2160",
  "fps": 60,
  "duration": 60,
  "ftp_endpoint": "127.0.0.1",
  "ftp_user": "TestUser",
  "ftp_password": "Xingjin***",
  "ftp_host": "127.0.0.1",
  "ftp_port": 21,
  "upload_directory": "/uploads"
}
```

Replace `<Your FTPS Endpoint>`, `<Your FTP Username>`, `<Your FTP Password>`, and `<Your FTP Upload Directory>` with your actual FileZilla server details.

### Usage

1. Open a terminal or command prompt and navigate to the project directory.
2. Run the following command to start the Test Data Generation Program:
   ```
   go run DataGenerator.go
   ```
3. The program will read the configuration from the `configuration.json` file and initiate the data generation process.
4. The generated video stream will include timestamps, and still images will be captured at the specified intervals.
5. The captured images will be securely uploaded to the FileZilla server using FTPS.

### Troubleshooting

If you encounter any issues during installation, configuration, or usage of the program, refer to the project's documentation or seek support from the development team.

## Contributing

Contributions to the Test Data Generation Program are welcome! To contribute, please follow these steps:

1. Fork the repository and create a new branch for your contribution.
2. Make your changes and ensure that the code adheres to the project's coding conventions.
3. Submit a pull request, describing the purpose and scope of your changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you need any assistance or have any questions, please contact our support team at [niyi.adebayo@remobi.co]. We are here to help!

## Disclaimer

In a production environment, it is important to set `InsecureSkipVerify` in the `tls.Config` to `false` for proper verification of the server's certificate chain and hostname. The current setting of `true` is intended for testing purposes only and should not be used in production deployments.

## Acknowledgments

We would like to acknowledge the contributions of the open-source community and the developers of the libraries and tools used in this project. Their efforts have made this Test Data Generation Program possible.
```

When you use this README in your project repository or a Markdown editor, it should display the correct headings and table of contents. 
 
 
 
 
 
 
 
 
