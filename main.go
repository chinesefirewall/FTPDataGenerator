package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Resolution  string `json:"resolution"`
	FPS         int    `json:"fps"`
	Duration    int    `json:"duration"`
	FTPUser     string `json:"ftp_user"`
	FTPPassword string `json:"ftp_password"`
	FTPHost     string `json:"ftp_host"`
	FTPPort     int    `json:"ftp_port"`
	OutputDir   string `json:"output_dir"`

	TestVideoPath     string `json:"test_video_path"`
	SnapshotOutputDir string `json:"snapshot_output_dir"`
	VideoOutputDir    string `json:"video_output_dir"`

	CsvOutputFile string `json:"csv_output_file"`

	Interval      int `json:"interval"`
	MaxRetries    int `json:"max_retries"`
	RetryInterval int `json:"retry_interval"`
	FTPConn       *ftp.ServerConn
}

// main is the primary entry point for the program. It handles the program's primary logic,
// including reading the configuration, creating the necessary directories, establishing
// FTPS connections, managing goroutines to create a test video and its snapshots, and
// generating metadata. It also coordinates the upload of the snapshots and metadata to
// the FTPS server. The function is designed to clean up resources and exit when all tasks
// have completed or upon encountering a fatal error.
func main() {
	// Read configuration from the JSON file
	config, err := readConfig("configuration.json")
	if err != nil {
		log.Fatalf("Failed to create output directory line 50: %v", err)
	}

	// Schedule cleanup to run when main function returns.
	//defer cleanup(config)

	// Create output directory if it doesn't exist.
	fmt.Println("Output Directory:", config.OutputDir)

	err = createDirectory(config.OutputDir)
	if err != nil {
		// If the output directory cannot be created, the program logs the error and exits.
		log.Fatalf("Failed to create output directory line 60: %v", err)
	}

	// Establish FTPS connection.
	err = establishFTPConnection(&config)
	if err != nil {
		// If the FTPS connection cannot be established, the program logs the error and decides.
		// whether to terminate or continue based on your logic.
		log.Printf("Failed to establish FTPS connection: %v", err)
		os.Exit(1)

	}
	// Create channels to communicate between goroutines.
	testVideoDone := make(chan bool)
	snapshotsDone := make(chan bool)
	//metadataDone := make(chan bool)

	// Generate a test video with timestamp concurrently.
	go func() {
		generateTestVideo(config)
		testVideoDone <- true
	}()

	go func() {
		// Wait for test video to complete before generating snapshots.
		<-testVideoDone
		generateSnapshots(config)
		snapshotsDone <- true
	}()

	go func() {
		// Wait for snapshots to be generated before generating metadata
		<-snapshotsDone
		generateMetadata(config)
	}()

	// A WaitGroup waits for a collection of goroutines to finish.
	var wg sync.WaitGroup

	// Upload snapshots and metadata to FTPS concurrently
	wg.Add(2)
	go func() {
		defer wg.Done()
		uploadSnapshots(&config)
	}()

	go func() {
		defer wg.Done()
		uploadMetadata(&config)
	}()

	wg.Wait() // Wait for all uploads to complete

	// Wait for the specified duration before stopping the generator
	time.Sleep(time.Second * time.Duration(config.Duration))

	// Program complete, print message and exit
	log.Println("Program complete and exiting")
}

// readConfig reads the configuration from the provided JSON file.
func readConfig(file string) (Config, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return Config{}, err
	}
	//defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

// createDirectory checks if the directory exists and creates it if it doesn't.
func createDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	return nil
}

func cleanup(config Config) {
	log.Println("Performing cleanup...")
	err := os.RemoveAll(config.OutputDir)
	if err != nil {
		log.Printf("Failed to clean up output directory: %v", err)
	}
	log.Println("Cleanup complete.")
}

// generateTestVideo generates a test video with timestamp.
func generateTestVideo(config Config) {
	log.Println("Generating test video...")
	var videoCmd = exec.Command("ffmpeg", "-f", "lavfi", "-i",
		fmt.Sprintf("testsrc=duration=%d:size=%s:rate=%d", config.Duration, config.Resolution, config.FPS),
		"-vf", fmt.Sprintf("drawtext=fontfile='/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf':text='%%{localtime}':x=(w-tw)/2:y=h-(2*lh):fontcolor=white:fontsize=12:box=1:boxcolor=black@0.5"), config.TestVideoPath)
	err := videoCmd.Run()
	if err != nil {
		log.Printf("Failed to generate test video: %v", err)
	}
	log.Println("Test video generation completed.")
}

// generateSnapshots generates snapshots from the test video at regular intervals.
func generateSnapshots(config Config) {
	log.Println("Generating snapshots...")

	// Before we start generating snapshots, we want to make sure that the directory
	// where we're going to save the snapshots exists. The os.MkdirAll function
	// creates a directory named by the path 'snapshotOutputDir', along with any necessary parents,
	// and returns nil if successful. The permission bits perm (before umask) are set for
	// all directories that MkdirAll creates. If the directory already exists, MkdirAll does nothing
	// and returns nil. In this case, we're setting the permissions to 0777, which means
	// everyone can read, write, and execute.
	// and later we can set the permissions 0755 (which means the owner can read,
	// write, and execute, while others can read and execute

	err := os.MkdirAll(config.SnapshotOutputDir, 0777)
	if err != nil {
		// If an error occurred while trying to create the directory, we log the error
		// and exit the function.
		log.Printf("Failed to create directory '%s': %v", config.SnapshotOutputDir, err)

		return
	}

	// We're using the ffmpeg tool to generate snapshots from the test video.
	// The snapshots are saved in the 'snapshotOutputDir' directory, with the filename
	// formatted as "snapshot%03d.jpg".

	// The ffmpeg command is executed using the exec.Command function, which creates
	snapshotCmd := exec.Command("ffmpeg", "-i", config.TestVideoPath, "-vf", fmt.Sprintf("fps=1/%d", config.Interval), filepath.Join(config.SnapshotOutputDir, "snapshot%03d.jpg"))

	// Run the command and wait for it to finish.
	err = snapshotCmd.Run()
	if err != nil {
		// If an error occurred while running the ffmpeg command, we log the error.
		log.Printf("Failed to generate snapshots: %v", err)
	}
	// Finally, we log that the snapshot generation has completed.
	log.Println("Snapshot generation completed.")
}

// generateMetadata generates a metadata.csv file with the names and creation times of the snapshot files.
func generateMetadata(config Config) {
	log.Println("Generating metadata...")

	// Retrieve snapshot files.
	snapshotFiles, err := filepath.Glob(filepath.Join(config.SnapshotOutputDir, "snapshot*.jpg"))
	if err != nil {
		log.Printf("Failed to retrieve snapshot files: %v", err)
		return
	}

	if len(snapshotFiles) == 0 {
		log.Println("Warning: No snapshot files found.")
		return // Don't proceed with generating metadata if there are no snapshots
	}

	// Prepare metadata records.
	var records [][]string
	records = append(records, []string{"Filename", "Creation Time"}) // CSV header
	for _, file := range snapshotFiles {
		fileInfo, err := os.Stat(file)
		if err != nil {
			log.Printf("Failed to retrieve file info for '%s': %v", file, err)
			continue
		}
		records = append(records, []string{filepath.Base(file), fileInfo.ModTime().String()})
	}

	// Create and write to metadata.csv
	file, err := os.Create(config.CsvOutputFile)
	if err != nil {
		log.Printf("Failed to create metadata file: %v", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close metadata file: %v", err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.WriteAll(records) // Write all records to the CSV file
	if err != nil {
		log.Printf("Failed to write to metadata file: %v", err)
		return
	}

	log.Println("Metadata generation completed.")
}

func uploadFile(config *Config, sourceFile string, targetFile string) (err error) {
	file, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	// defer closure that checks the returned error from file.Close()
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			// If there was an error during the file upload and closing the file, we want to keep the original error.
			if err == nil {
				err = closeErr
			} else {
				log.Printf("Failed to close the file: %v", closeErr)
			}
		}
	}()

	err = config.FTPConn.Stor(targetFile, file)
	if err != nil {
		return err
	}
	return nil
}

// establishFTPConnection establishes a connection to the FTP server.
func establishFTPConnection(config *Config) error {
	addr := fmt.Sprintf("%s:%d", config.FTPHost, config.FTPPort)

	for i := 0; i < config.MaxRetries; i++ {
		c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			log.Printf("Failed to establish FTP connection, attempt %d/%d: %v", i+1, config.MaxRetries, err)
			time.Sleep(time.Duration(config.RetryInterval))
			continue
		}

		err = c.Login(config.FTPUser, config.FTPPassword)
		if err != nil {
			log.Printf("Failed to authenticate, attempt %d/%d: %v", i+1, config.MaxRetries, err)
			time.Sleep(time.Duration(config.RetryInterval))
			continue
		}

		config.FTPConn = c
		return nil
	}

	return fmt.Errorf("failed to establish FTP connection after %d attempts", config.MaxRetries)
}

func uploadSnapshots(config *Config) {
	log.Println("Uploading snapshots to FTPS...")
	snapshotFiles, err := filepath.Glob(filepath.Join(config.SnapshotOutputDir, "snapshot*.jpg"))
	if err != nil {
		log.Printf("Failed to retrieve snapshot files: %v", err)
		return
	}

	for _, file := range snapshotFiles {
		err = uploadFile(config, file, filepath.Join(config.OutputDir, filepath.Base(file)))
		if err != nil {
			log.Printf("Failed to upload snapshot file '%s': %v", file, err)
		} else {
			log.Printf("Uploaded snapshot file '%s'", file)
		}

		time.Sleep(time.Millisecond * time.Duration(config.Interval))
	}

	log.Println("Snapshot upload completed.")
}

// uploadMetadata uploads metadata to the FTPS.
func uploadMetadata(config *Config) {
	log.Println("Uploading metadata to FTPS...")
	err := uploadFile(config, config.CsvOutputFile, filepath.Join(config.OutputDir, "metadata.csv"))
	if err != nil {
		log.Printf("Failed to upload metadata: %v", err)
	} else {
		log.Println("Metadata upload completed.")
	}
}
