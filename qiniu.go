package qiniu

import (
	qiniu "qiniupkg.com/api.v7/kodo"

	"github.com/docker/distribution/context"

	"fmt"
	"github.com/docker/distribution/registry/storage/driver/base"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/factory"
	"net/http"

	"io"
	"strconv"

	"bytes"
	"qiniupkg.com/api.v7/kodocli"
	"encoding/base64"
)

const driverName  = "qiniu"

const (
	paramAccountName = "accesskey"
	paramAccountKey  = "secretkey"
	paramBucket      = "bucket"
	paramDomain      = "domain"
	paramIsPrivate   = "isprivate"
	paramZone	 = "zone"

	paramRSHost      = "rshost"
	paramRSFHost     = "rsfhost"
	paramAPIHost     = "apihost"
	paramScheme      = "scheme"
	paramIoHost      = "iohost"
	paramUphosts     = "uphosts"

	maxChunkSize     = 4 * 1024 * 1024
	blockSize        = 4 * 1024 * 1024
	chunkSize        = 1 * 1024 * 1024   //1M
)




//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	qiniu.Config
	Bucket string
	Zone int
	Domain string
	IsPrivate bool
}

type baseEmbed struct{ base.Base }

// Driver is a storagedriver.StorageDriver implementation backed by
// Microsoft Azure Blob Storage Service.
type Driver struct{ baseEmbed }



func init() {
	factory.Register(driverName, &qiniuDriverFactory{})
}


type qiniuDriverFactory struct{}

func (factory *qiniuDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	Client *qiniu.Client
	Uploader kodocli.Uploader
	Bucket qiniu.Bucket
	Config DriverParameters

}

/**
 * FromParameters constructs a new Driver with a given parameters map
 * Required parameters:
 * accesskey
 * rshost
 * rsfhost
 * apihost
 * encrypt
 */
func FromParameters(parameters map[string]interface{}) (*Driver,error) {
	// Providing no values for these is valid in case the user is authenticating

	accountName, ok := parameters[paramAccountName]
	if !ok || fmt.Sprint(accountName) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramAccountName)
	}

	accountKey, ok := parameters[paramAccountKey]
	if !ok || fmt.Sprint(accountKey) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramAccountKey)
	}

	bucket, ok := parameters[paramBucket]
	if !ok || fmt.Sprint(accountKey) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramBucket)
	}

	domain, ok := parameters[paramDomain]
	if !ok || fmt.Sprint(accountKey) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramDomain)
	}

	isPrivate, ok := parameters[paramIsPrivate]
	if !ok || fmt.Sprint(accountKey) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramIsPrivate)
	}

	params := DriverParameters{}
	params.AccessKey = fmt.Sprint(accountName)
	params.SecretKey = fmt.Sprint(accountKey)
	params.Bucket = fmt.Sprint(bucket)
	params.Domain = fmt.Sprint(domain)
	params.IsPrivate = fmt.Sprint(isPrivate) == "true"

	fmt.Print(params)
	return nil,nil
}

func New(params DriverParameters) (*Driver, error)  {
	qiniuConfig := qiniu.Config{
		AccessKey: params.AccessKey,
		SecretKey: params.SecretKey,
	}
	client := qiniu.New(params.Zone,&qiniuConfig)
	bucket := client.Bucket(params.Bucket)
	d := &driver{
		Client: client,
		Bucket: bucket,
		Config: params,
		Uploader: kodocli.NewUploader(params.Zone, nil),
	}

	return &Driver{
		baseEmbed: baseEmbed{
			Base: base.Base{
				StorageDriver: d,
			},
		},
	}, nil
}

// Implement the storagedriver.StorageDriver interface

func (d *driver) Name() string  {
	return driverName
}

// GetContent retrieves the content stored at "path" as a []byte.
func (d *driver) GetContent(ctx context.Context, path string) ([]byte, error) {
	baseUrl := d.getBaseUrl(path)
	res, err := http.Get(baseUrl)
	content := make([]byte,res.ContentLength)
	length,err := res.Body.Read(content)
	if err != nil || int64(length) != res.ContentLength {
		return nil, nil
	}
	return content, nil
}


// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	if len(contents) > maxChunkSize { // max size for block data uploaded
		return fmt.Errorf("uploading %d bytes with PutContent is not supported; limit: %d bytes", len(contents), maxChunkSize)
	}
	reader := bytes.NewReader(contents)
	return d.Bucket.Put(ctx, nil, path, reader, int64(len(contents)), nil)
}


// Reader retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	baseUrl := d.getBaseUrl(path)

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl, nil)
	req.Header.Add("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	return resp.Body,err
}



// Writer returns a FileWriter which will store the content written to it
// at the location designated by "path" after the call to Commit.
func (d *driver) Writer(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error) {
	exist := d.fileExists(path)

	if exist {
		if append {
			//TODO append data to an exist file
		} else {
			err := d.Bucket.Delete(ctx,path)
			if err != nil {
				return nil, err
			}
		}
	}

	return writer{
		driver: d,
		key: path,
		blocks: []block{},
		size: 0,
		closed: false,
		committed: false,
		cancelled: false,
	},nil

}


//retrieve a url from a path
func (d *driver) getBaseUrl(path string) string {
	baseUrl := qiniu.MakeBaseUrl(d.Config.Domain, path)
	if d.Config.IsPrivate {
		baseUrl = d.Client.MakePrivateUrl(baseUrl)
	}
	return baseUrl
}

func (d *driver) fileExists(path string) bool  {
	_,err :=d.Bucket.Stat(path)
	if err {
		return true
	}
	return false
}

type block struct {
	size int64
	data []byte
	lastCtx string
	finished bool
}




// writer attempts to upload parts to S3 in a buffered fashion where the last
// part is at least as large as the chunksize, so the multipart upload could be
// cleanly resumed in the future. This is violated if Close is called after less
// than a full chunk is written.
type writer struct {
	driver      *driver
	key         string
	blocks      []block
	size        int64
	readyPart   []byte
	pendingPart []byte
	closed      bool
	committed   bool
	cancelled   bool
}

func (d *driver) newWriter(key string, path string) storagedriver.FileWriter {

	return &writer{
		driver: d,
		key:    path,
		size:   0,
		blocks: []block{},
	}
}

func (w *writer) Write(p []byte) (int, error) {
	if w.closed {
		return 0, fmt.Errorf("already closed")
	} else if w.committed {
		return 0, fmt.Errorf("already committed")
	} else if w.cancelled {
		return 0, fmt.Errorf("already cancelled")
	} else if len(p) == 0 {
		return 0, fmt.Errorf("empty data")
	}

	w.append(p)
	w.flushBlock()

	return len(p), nil
}

func (w *writer) Size() int64 {
	return w.size
}

func (w *writer) Close() error {
	if w.closed {
		return fmt.Errorf("already closed")
	}
	err := w.flushBlock()
	if err != nil {
		error("flush block error")
	}
	return w.mkfile()
}

func (w *writer) Cancel() error {
	if w.closed {
		return fmt.Errorf("already closed")
	} else if w.committed {
		return fmt.Errorf("already committed")
	}
	w.cancelled = true
	return nil
}

func (w *writer) Commit() error {
	if w.closed {
		return fmt.Errorf("already closed")
	} else if w.committed {
		return fmt.Errorf("already committed")
	} else if w.cancelled {
		return fmt.Errorf("already cancelled")
	}

	return w.flushBlock()
}

// flush buffers to write
// Only not full block will be flushed
func (w *writer) flushBlock() error {

	for i:=len(w.blocks);i>=0;i-- {
		if !w.blocks[i].finished && w.blocks[i].size == blockSize {
			w.uploadBlock(&w.blocks[i])
		}
	}

	return nil

}


func (w *writer) append(data[]byte)  {
	length := len(w.blocks)

	//complement the last block
	if length > 0 &&
		w.blocks[length-1].size < blockSize &&
		! w.blocks[length-1].finished {
		last := w.blocks[length-1]
		idx := min(blockSize - last.size,len(data))
		last.data = append(last.data, data[:idx])
		data = data[idx:]
	}

	for len(data) > 0 {
		sz := min(blockSize, len(data))
		append(w.blocks,block{
			size: sz,
			data: data[:sz],
			finished: false,
		})
		data = data[sz:]
	}
}


func (w *writer) uploadBlock(blk *block)  {
	w.driver.Uploader.Conn.Call()
	url := w.driver.Client.UpHosts[0] + "/mkblk/" + strconv.Itoa(blk.size)
	p := blk.data
	idx := min(blk.size, chunkSize)

	//creat a block
	firstChunk := p[:idx]
	p = p[idx:]
	body := bytes.NewReader(firstChunk)
	nextChunkInfo, err := request("POST", url, "application/octet-stream", body, idx)
	if err != nil {
		return nil, err
	}

	//upload chunk
	for len(p)>0  {
		idx = min(len(p), chunkSize)
		chunk := p[:idx]
		p = p[idx:]
		body = bytes.NewReader(chunk)

		url = nextChunkInfo["host"] + "/bput/" + nextChunkInfo["ctx"] + "/" + strconv.FormatUint(uint64(nextChunkInfo["offset"]), 10)

		nextChunkInfo, err =request("POST", url, "application/octet-stream", body, idx)
		if err != nil {
			return nil, err
		}
	}

	blk.finished = true
	blk.lastCtx = nextChunkInfo["ctx"]

}

func (w *writer) mkfile() error {
	blocks := w.blocks
	blockNum := len(blocks)
	if blockNum == 0 {
		return fmt.Errorf("empty blocks")
	}

	if ! blocks[blockNum-1].finished {
		w.uploadBlock(&blocks[blockNum-1])
	}

	var content string
	for i:= 0; i< blockNum; i++ {
		content += "," + blocks[i].lastCtx
	}
	content = content[1:]


	url := w.driver.Client.UpHosts[0] + "/mkfile/" +
		strconv.FormatInt(w.size, 10) +
		"/key/" + base64.URLEncoding.EncodeToString([]byte(w.key))
	_, err := request("POST", url, "application/octet-stream", bytes.NewReader(content), len(content))
	return err
}