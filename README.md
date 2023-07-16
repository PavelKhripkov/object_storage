# OBJECT STORAGE

#### This is concept showing how basic functionality of an object storage can be implemented.

### Key features:
1. Fast uploading. First loads item directly on the server, then splits it into chunks and passes to the remote file servers.
2. Immediate downloading. No need to wait for chunks to be downloaded from remoter storages. On download request, stream is created and chunks can be consumed directly from remote storages.
3. Different file servers can be used to store items (API, SSH, FTP, etc).

### Testing

Test module implemented in **api_test/main.go**. Currently, it must be configured directly in the code and run manually. Test program creates randomly generated file of specified size, uploads it to the storage, then downloads and compares MD5 hash sum.

#### What is not implemented yet:
- Unit tests
- API comprehensive tests
- Migrations
- File server communication except SSH
- Storage layer except SQLite
- Some basic features like entity removing