// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: idl/cloudstorage/file.proto

package file

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	FileService_Upload_FullMethodName            = "/file.FileService/Upload"
	FileService_CreateFileStore_FullMethodName   = "/file.FileService/CreateFileStore"
	FileService_CreateFolder_FullMethodName      = "/file.FileService/CreateFolder"
	FileService_ListFolder_FullMethodName        = "/file.FileService/ListFolder"
	FileService_GetFile_FullMethodName           = "/file.FileService/GetFile"
	FileService_Download_FullMethodName          = "/file.FileService/Download"
	FileService_DownloadStream_FullMethodName    = "/file.FileService/DownloadStream"
	FileService_MoveFolder_FullMethodName        = "/file.FileService/MoveFolder"
	FileService_MoveFile_FullMethodName          = "/file.FileService/MoveFile"
	FileService_DeleteFile_FullMethodName        = "/file.FileService/DeleteFile"
	FileService_DeleteFolder_FullMethodName      = "/file.FileService/DeleteFolder"
	FileService_Search_FullMethodName            = "/file.FileService/Search"
	FileService_Preview_FullMethodName           = "/file.FileService/Preview"
	FileService_DownloadTask_FullMethodName      = "/file.FileService/DownloadTask"
	FileService_GetDownloadTask_FullMethodName   = "/file.FileService/GetDownloadTask"
	FileService_ResumeDownload_FullMethodName    = "/file.FileService/ResumeDownload"
	FileService_UploadChunkStream_FullMethodName = "/file.FileService/UploadChunkStream"
	FileService_CreateShareLink_FullMethodName   = "/file.FileService/CreateShareLink"
	FileService_SaveToMyDrive_FullMethodName     = "/file.FileService/SaveToMyDrive"
	FileService_GetUserFileStore_FullMethodName  = "/file.FileService/GetUserFileStore"
	FileService_UpdateFile_FullMethodName        = "/file.FileService/UpdateFile"
)

// FileServiceClient is the client API for FileService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileServiceClient interface {
	Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadResponse, error)
	CreateFileStore(ctx context.Context, in *CreateFileStoreRequest, opts ...grpc.CallOption) (*CreateFileStoreResponse, error)
	CreateFolder(ctx context.Context, in *CreateFolderRequest, opts ...grpc.CallOption) (*CreateFolderResponse, error)
	ListFolder(ctx context.Context, in *ListFolderRequest, opts ...grpc.CallOption) (*ListFolderResponse, error)
	GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (*GetFileResponse, error)
	Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error)
	DownloadStream(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[DownloadStreamResponse], error)
	MoveFolder(ctx context.Context, in *MoveFolderRequest, opts ...grpc.CallOption) (*MoveFolderResponse, error)
	MoveFile(ctx context.Context, in *MoveFileRequest, opts ...grpc.CallOption) (*MoveFileResponse, error)
	DeleteFile(ctx context.Context, in *DeleteFileRequest, opts ...grpc.CallOption) (*DeleteFileResponse, error)
	DeleteFolder(ctx context.Context, in *DeleteFolderRequest, opts ...grpc.CallOption) (*DeleteFolderResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
	Preview(ctx context.Context, in *PreviewRequest, opts ...grpc.CallOption) (*PreviewResponse, error)
	DownloadTask(ctx context.Context, in *DownloadTaskRequest, opts ...grpc.CallOption) (*DownloadTaskResponse, error)
	GetDownloadTask(ctx context.Context, in *GetDownloadTaskRequest, opts ...grpc.CallOption) (*GetDownloadTaskResponse, error)
	ResumeDownload(ctx context.Context, in *ResumeDownloadRequest, opts ...grpc.CallOption) (*ResumeDownloadResponse, error)
	UploadChunkStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[UploadChunkRequest, UploadChunkResponse], error)
	CreateShareLink(ctx context.Context, in *CreateShareLinkRequest, opts ...grpc.CallOption) (*CreateShareLinkResponse, error)
	SaveToMyDrive(ctx context.Context, in *SaveToMyDriveRequest, opts ...grpc.CallOption) (*SaveToMyDriveResponse, error)
	GetUserFileStore(ctx context.Context, in *GetUserFileStoreRequest, opts ...grpc.CallOption) (*GetUserFileStoreResponse, error)
	UpdateFile(ctx context.Context, in *UpdateFileRequest, opts ...grpc.CallOption) (*UpdateFileResponse, error)
}

type fileServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFileServiceClient(cc grpc.ClientConnInterface) FileServiceClient {
	return &fileServiceClient{cc}
}

func (c *fileServiceClient) Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UploadResponse)
	err := c.cc.Invoke(ctx, FileService_Upload_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) CreateFileStore(ctx context.Context, in *CreateFileStoreRequest, opts ...grpc.CallOption) (*CreateFileStoreResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateFileStoreResponse)
	err := c.cc.Invoke(ctx, FileService_CreateFileStore_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) CreateFolder(ctx context.Context, in *CreateFolderRequest, opts ...grpc.CallOption) (*CreateFolderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateFolderResponse)
	err := c.cc.Invoke(ctx, FileService_CreateFolder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) ListFolder(ctx context.Context, in *ListFolderRequest, opts ...grpc.CallOption) (*ListFolderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListFolderResponse)
	err := c.cc.Invoke(ctx, FileService_ListFolder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (*GetFileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetFileResponse)
	err := c.cc.Invoke(ctx, FileService_GetFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DownloadResponse)
	err := c.cc.Invoke(ctx, FileService_Download_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) DownloadStream(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[DownloadStreamResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &FileService_ServiceDesc.Streams[0], FileService_DownloadStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[DownloadRequest, DownloadStreamResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileService_DownloadStreamClient = grpc.ServerStreamingClient[DownloadStreamResponse]

func (c *fileServiceClient) MoveFolder(ctx context.Context, in *MoveFolderRequest, opts ...grpc.CallOption) (*MoveFolderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MoveFolderResponse)
	err := c.cc.Invoke(ctx, FileService_MoveFolder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) MoveFile(ctx context.Context, in *MoveFileRequest, opts ...grpc.CallOption) (*MoveFileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MoveFileResponse)
	err := c.cc.Invoke(ctx, FileService_MoveFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) DeleteFile(ctx context.Context, in *DeleteFileRequest, opts ...grpc.CallOption) (*DeleteFileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteFileResponse)
	err := c.cc.Invoke(ctx, FileService_DeleteFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) DeleteFolder(ctx context.Context, in *DeleteFolderRequest, opts ...grpc.CallOption) (*DeleteFolderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteFolderResponse)
	err := c.cc.Invoke(ctx, FileService_DeleteFolder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SearchResponse)
	err := c.cc.Invoke(ctx, FileService_Search_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) Preview(ctx context.Context, in *PreviewRequest, opts ...grpc.CallOption) (*PreviewResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PreviewResponse)
	err := c.cc.Invoke(ctx, FileService_Preview_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) DownloadTask(ctx context.Context, in *DownloadTaskRequest, opts ...grpc.CallOption) (*DownloadTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DownloadTaskResponse)
	err := c.cc.Invoke(ctx, FileService_DownloadTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) GetDownloadTask(ctx context.Context, in *GetDownloadTaskRequest, opts ...grpc.CallOption) (*GetDownloadTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetDownloadTaskResponse)
	err := c.cc.Invoke(ctx, FileService_GetDownloadTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) ResumeDownload(ctx context.Context, in *ResumeDownloadRequest, opts ...grpc.CallOption) (*ResumeDownloadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResumeDownloadResponse)
	err := c.cc.Invoke(ctx, FileService_ResumeDownload_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) UploadChunkStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[UploadChunkRequest, UploadChunkResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &FileService_ServiceDesc.Streams[1], FileService_UploadChunkStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[UploadChunkRequest, UploadChunkResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileService_UploadChunkStreamClient = grpc.ClientStreamingClient[UploadChunkRequest, UploadChunkResponse]

func (c *fileServiceClient) CreateShareLink(ctx context.Context, in *CreateShareLinkRequest, opts ...grpc.CallOption) (*CreateShareLinkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShareLinkResponse)
	err := c.cc.Invoke(ctx, FileService_CreateShareLink_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) SaveToMyDrive(ctx context.Context, in *SaveToMyDriveRequest, opts ...grpc.CallOption) (*SaveToMyDriveResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SaveToMyDriveResponse)
	err := c.cc.Invoke(ctx, FileService_SaveToMyDrive_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) GetUserFileStore(ctx context.Context, in *GetUserFileStoreRequest, opts ...grpc.CallOption) (*GetUserFileStoreResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserFileStoreResponse)
	err := c.cc.Invoke(ctx, FileService_GetUserFileStore_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) UpdateFile(ctx context.Context, in *UpdateFileRequest, opts ...grpc.CallOption) (*UpdateFileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateFileResponse)
	err := c.cc.Invoke(ctx, FileService_UpdateFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileServiceServer is the server API for FileService service.
// All implementations must embed UnimplementedFileServiceServer
// for forward compatibility.
type FileServiceServer interface {
	Upload(context.Context, *UploadRequest) (*UploadResponse, error)
	CreateFileStore(context.Context, *CreateFileStoreRequest) (*CreateFileStoreResponse, error)
	CreateFolder(context.Context, *CreateFolderRequest) (*CreateFolderResponse, error)
	ListFolder(context.Context, *ListFolderRequest) (*ListFolderResponse, error)
	GetFile(context.Context, *GetFileRequest) (*GetFileResponse, error)
	Download(context.Context, *DownloadRequest) (*DownloadResponse, error)
	DownloadStream(*DownloadRequest, grpc.ServerStreamingServer[DownloadStreamResponse]) error
	MoveFolder(context.Context, *MoveFolderRequest) (*MoveFolderResponse, error)
	MoveFile(context.Context, *MoveFileRequest) (*MoveFileResponse, error)
	DeleteFile(context.Context, *DeleteFileRequest) (*DeleteFileResponse, error)
	DeleteFolder(context.Context, *DeleteFolderRequest) (*DeleteFolderResponse, error)
	Search(context.Context, *SearchRequest) (*SearchResponse, error)
	Preview(context.Context, *PreviewRequest) (*PreviewResponse, error)
	DownloadTask(context.Context, *DownloadTaskRequest) (*DownloadTaskResponse, error)
	GetDownloadTask(context.Context, *GetDownloadTaskRequest) (*GetDownloadTaskResponse, error)
	ResumeDownload(context.Context, *ResumeDownloadRequest) (*ResumeDownloadResponse, error)
	UploadChunkStream(grpc.ClientStreamingServer[UploadChunkRequest, UploadChunkResponse]) error
	CreateShareLink(context.Context, *CreateShareLinkRequest) (*CreateShareLinkResponse, error)
	SaveToMyDrive(context.Context, *SaveToMyDriveRequest) (*SaveToMyDriveResponse, error)
	GetUserFileStore(context.Context, *GetUserFileStoreRequest) (*GetUserFileStoreResponse, error)
	UpdateFile(context.Context, *UpdateFileRequest) (*UpdateFileResponse, error)
	mustEmbedUnimplementedFileServiceServer()
}

// UnimplementedFileServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedFileServiceServer struct{}

func (UnimplementedFileServiceServer) Upload(context.Context, *UploadRequest) (*UploadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Upload not implemented")
}
func (UnimplementedFileServiceServer) CreateFileStore(context.Context, *CreateFileStoreRequest) (*CreateFileStoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFileStore not implemented")
}
func (UnimplementedFileServiceServer) CreateFolder(context.Context, *CreateFolderRequest) (*CreateFolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFolder not implemented")
}
func (UnimplementedFileServiceServer) ListFolder(context.Context, *ListFolderRequest) (*ListFolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFolder not implemented")
}
func (UnimplementedFileServiceServer) GetFile(context.Context, *GetFileRequest) (*GetFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFile not implemented")
}
func (UnimplementedFileServiceServer) Download(context.Context, *DownloadRequest) (*DownloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Download not implemented")
}
func (UnimplementedFileServiceServer) DownloadStream(*DownloadRequest, grpc.ServerStreamingServer[DownloadStreamResponse]) error {
	return status.Errorf(codes.Unimplemented, "method DownloadStream not implemented")
}
func (UnimplementedFileServiceServer) MoveFolder(context.Context, *MoveFolderRequest) (*MoveFolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MoveFolder not implemented")
}
func (UnimplementedFileServiceServer) MoveFile(context.Context, *MoveFileRequest) (*MoveFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MoveFile not implemented")
}
func (UnimplementedFileServiceServer) DeleteFile(context.Context, *DeleteFileRequest) (*DeleteFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFile not implemented")
}
func (UnimplementedFileServiceServer) DeleteFolder(context.Context, *DeleteFolderRequest) (*DeleteFolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFolder not implemented")
}
func (UnimplementedFileServiceServer) Search(context.Context, *SearchRequest) (*SearchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedFileServiceServer) Preview(context.Context, *PreviewRequest) (*PreviewResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Preview not implemented")
}
func (UnimplementedFileServiceServer) DownloadTask(context.Context, *DownloadTaskRequest) (*DownloadTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadTask not implemented")
}
func (UnimplementedFileServiceServer) GetDownloadTask(context.Context, *GetDownloadTaskRequest) (*GetDownloadTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDownloadTask not implemented")
}
func (UnimplementedFileServiceServer) ResumeDownload(context.Context, *ResumeDownloadRequest) (*ResumeDownloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResumeDownload not implemented")
}
func (UnimplementedFileServiceServer) UploadChunkStream(grpc.ClientStreamingServer[UploadChunkRequest, UploadChunkResponse]) error {
	return status.Errorf(codes.Unimplemented, "method UploadChunkStream not implemented")
}
func (UnimplementedFileServiceServer) CreateShareLink(context.Context, *CreateShareLinkRequest) (*CreateShareLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShareLink not implemented")
}
func (UnimplementedFileServiceServer) SaveToMyDrive(context.Context, *SaveToMyDriveRequest) (*SaveToMyDriveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveToMyDrive not implemented")
}
func (UnimplementedFileServiceServer) GetUserFileStore(context.Context, *GetUserFileStoreRequest) (*GetUserFileStoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserFileStore not implemented")
}
func (UnimplementedFileServiceServer) UpdateFile(context.Context, *UpdateFileRequest) (*UpdateFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFile not implemented")
}
func (UnimplementedFileServiceServer) mustEmbedUnimplementedFileServiceServer() {}
func (UnimplementedFileServiceServer) testEmbeddedByValue()                     {}

// UnsafeFileServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileServiceServer will
// result in compilation errors.
type UnsafeFileServiceServer interface {
	mustEmbedUnimplementedFileServiceServer()
}

func RegisterFileServiceServer(s grpc.ServiceRegistrar, srv FileServiceServer) {
	// If the following call pancis, it indicates UnimplementedFileServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&FileService_ServiceDesc, srv)
}

func _FileService_Upload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).Upload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_Upload_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).Upload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_CreateFileStore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFileStoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).CreateFileStore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_CreateFileStore_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).CreateFileStore(ctx, req.(*CreateFileStoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_CreateFolder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFolderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).CreateFolder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_CreateFolder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).CreateFolder(ctx, req.(*CreateFolderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_ListFolder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFolderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).ListFolder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_ListFolder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).ListFolder(ctx, req.(*ListFolderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_GetFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).GetFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_GetFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).GetFile(ctx, req.(*GetFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_Download_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).Download(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_Download_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).Download(ctx, req.(*DownloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_DownloadStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileServiceServer).DownloadStream(m, &grpc.GenericServerStream[DownloadRequest, DownloadStreamResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileService_DownloadStreamServer = grpc.ServerStreamingServer[DownloadStreamResponse]

func _FileService_MoveFolder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveFolderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).MoveFolder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_MoveFolder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).MoveFolder(ctx, req.(*MoveFolderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_MoveFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).MoveFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_MoveFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).MoveFile(ctx, req.(*MoveFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_DeleteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).DeleteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_DeleteFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).DeleteFile(ctx, req.(*DeleteFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_DeleteFolder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFolderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).DeleteFolder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_DeleteFolder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).DeleteFolder(ctx, req.(*DeleteFolderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_Search_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_Preview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).Preview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_Preview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).Preview(ctx, req.(*PreviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_DownloadTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).DownloadTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_DownloadTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).DownloadTask(ctx, req.(*DownloadTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_GetDownloadTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDownloadTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).GetDownloadTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_GetDownloadTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).GetDownloadTask(ctx, req.(*GetDownloadTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_ResumeDownload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResumeDownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).ResumeDownload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_ResumeDownload_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).ResumeDownload(ctx, req.(*ResumeDownloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_UploadChunkStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FileServiceServer).UploadChunkStream(&grpc.GenericServerStream[UploadChunkRequest, UploadChunkResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileService_UploadChunkStreamServer = grpc.ClientStreamingServer[UploadChunkRequest, UploadChunkResponse]

func _FileService_CreateShareLink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShareLinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).CreateShareLink(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_CreateShareLink_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).CreateShareLink(ctx, req.(*CreateShareLinkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_SaveToMyDrive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveToMyDriveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).SaveToMyDrive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_SaveToMyDrive_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).SaveToMyDrive(ctx, req.(*SaveToMyDriveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_GetUserFileStore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserFileStoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).GetUserFileStore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_GetUserFileStore_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).GetUserFileStore(ctx, req.(*GetUserFileStoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_UpdateFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).UpdateFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_UpdateFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).UpdateFile(ctx, req.(*UpdateFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FileService_ServiceDesc is the grpc.ServiceDesc for FileService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FileService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "file.FileService",
	HandlerType: (*FileServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Upload",
			Handler:    _FileService_Upload_Handler,
		},
		{
			MethodName: "CreateFileStore",
			Handler:    _FileService_CreateFileStore_Handler,
		},
		{
			MethodName: "CreateFolder",
			Handler:    _FileService_CreateFolder_Handler,
		},
		{
			MethodName: "ListFolder",
			Handler:    _FileService_ListFolder_Handler,
		},
		{
			MethodName: "GetFile",
			Handler:    _FileService_GetFile_Handler,
		},
		{
			MethodName: "Download",
			Handler:    _FileService_Download_Handler,
		},
		{
			MethodName: "MoveFolder",
			Handler:    _FileService_MoveFolder_Handler,
		},
		{
			MethodName: "MoveFile",
			Handler:    _FileService_MoveFile_Handler,
		},
		{
			MethodName: "DeleteFile",
			Handler:    _FileService_DeleteFile_Handler,
		},
		{
			MethodName: "DeleteFolder",
			Handler:    _FileService_DeleteFolder_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _FileService_Search_Handler,
		},
		{
			MethodName: "Preview",
			Handler:    _FileService_Preview_Handler,
		},
		{
			MethodName: "DownloadTask",
			Handler:    _FileService_DownloadTask_Handler,
		},
		{
			MethodName: "GetDownloadTask",
			Handler:    _FileService_GetDownloadTask_Handler,
		},
		{
			MethodName: "ResumeDownload",
			Handler:    _FileService_ResumeDownload_Handler,
		},
		{
			MethodName: "CreateShareLink",
			Handler:    _FileService_CreateShareLink_Handler,
		},
		{
			MethodName: "SaveToMyDrive",
			Handler:    _FileService_SaveToMyDrive_Handler,
		},
		{
			MethodName: "GetUserFileStore",
			Handler:    _FileService_GetUserFileStore_Handler,
		},
		{
			MethodName: "UpdateFile",
			Handler:    _FileService_UpdateFile_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DownloadStream",
			Handler:       _FileService_DownloadStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UploadChunkStream",
			Handler:       _FileService_UploadChunkStream_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "idl/cloudstorage/file.proto",
}
