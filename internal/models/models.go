package models

// ConnectRequest 连接请求结构
type ConnectRequest struct {
	Endpoint string `json:"endpoint"`
	AK       string `json:"ak"`
	SK       string `json:"sk"`
	Bucket   string `json:"bucket"`
	SSL      bool   `json:"ssl"`
}

// ListRequest 列表请求结构
type ListRequest struct {
	ConnectRequest
	Prefix string `json:"prefix"`
	Limit  int    `json:"limit"`  // 分页大小
	Marker string `json:"marker"` // 分页标记
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"lastModified"`
}

// Response 通用响应
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ListResponse 列表响应
type ListResponse struct {
	Objects    []ObjectInfo `json:"objects"`
	Folders    []string     `json:"folders,omitempty"`    // 文件夹列表（CommonPrefixes）
	NextMarker string       `json:"nextMarker,omitempty"` // 下一页标记
	HasMore    bool         `json:"hasMore"`              // 是否有更多
}

// BucketInfo 存储桶信息
type BucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creationDate"`
}

// BucketsResponse 存储桶列表响应
type BucketsResponse struct {
	Success bool         `json:"success"`
	Buckets []BucketInfo `json:"buckets,omitempty"`
	Message string       `json:"message,omitempty"`
}

// DownloadRequest 下载请求结构
type DownloadRequest struct {
	ConnectRequest
	Key string `json:"key"`
}

// DeleteRequest 删除请求结构
type DeleteRequest struct {
	ConnectRequest
	Key string `json:"key"`
}

// MkdirRequest 创建文件夹请求结构
type MkdirRequest struct {
	ConnectRequest
	Folder string `json:"folder"`
}

// RmdirRequest 删除文件夹请求结构
type RmdirRequest struct {
	ConnectRequest
	Folder string `json:"folder"`
}

// CorsRequest CORS配置请求结构
type CorsRequest struct {
	ConnectRequest
	Config string `json:"config,omitempty"` // JSON格式的CORS配置（仅POST使用）
}

// CorsResponse CORS配置响应结构
type CorsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Config  string `json:"config,omitempty"` // JSON格式的CORS配置
}

// BatchDeleteRequest 批量删除请求结构
type BatchDeleteRequest struct {
	ConnectRequest
	Keys []string `json:"keys"`
}

// DeleteResult 单个删除结果
type DeleteResult struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BatchDeleteResponse 批量删除响应结构
type BatchDeleteResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Results []DeleteResult `json:"results,omitempty"`
}

// PreviewRequest 预览请求结构
type PreviewRequest struct {
	ConnectRequest
	Key string `json:"key"`
}

// PreviewResponse 预览响应结构
type PreviewResponse struct {
	Success     bool   `json:"success"`
	ContentType string `json:"contentType,omitempty"`
	Data        string `json:"data,omitempty"` // 文本内容或 base64 编码的数据
	URL         string `json:"url,omitempty"`  // 预签名 URL（大文件/视频）
	Message     string `json:"message,omitempty"`
}

// RenameRequest 重命名请求结构
type RenameRequest struct {
	ConnectRequest
	OldKey string `json:"oldKey"`
	NewKey string `json:"newKey"`
}

// MoveRequest 移动请求结构
type MoveRequest struct {
	ConnectRequest
	SrcKey string `json:"srcKey"`
	DstKey string `json:"dstKey"`
}

// CopyRequest 复制请求结构
type CopyRequest struct {
	ConnectRequest
	SrcKey string `json:"srcKey"`
	DstKey string `json:"dstKey"`
}
