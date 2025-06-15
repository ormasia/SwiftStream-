import requests
import json
import hashlib

BASE_URL = 'http://localhost:8080/api/oss'

# 计算文件MD5
def calculate_md5(data):
    return hashlib.md5(data).hexdigest()

# 1. 初始化上传任务
def init_upload(filename, filesize, filetype, chunk_size, file_md5=None):
    data = {
        'file_name': filename,
        'file_size': filesize,
        'file_type': filetype,
        'chunk_size': chunk_size
    }
    if file_md5:
        data['file_md5'] = file_md5
    
    resp = requests.post(f'{BASE_URL}/upload/init', json=data)
    print('Init:', resp.status_code, resp.text)
    resp.raise_for_status()
    
    response_data = resp.json()
    # 检查是否为秒传
    if response_data.get('chunkCount', 0) == 0:
        print(f"秒传成功！文件URL: {response_data['fileUrl']}")
        return response_data['uploadId'], 0 # 秒传时 chunkCount 为 0

    return response_data['uploadId'], response_data['chunkCount']

# 2. 上传分片
def upload_chunk(upload_id, chunk_index, data):
    files = {'chunk': (f'chunk{chunk_index}', data)}
    url = f'{BASE_URL}/upload/{upload_id}/chunk/{chunk_index}'
    resp = requests.post(url, files=files)
    print(f'Chunk {chunk_index}:', resp.status_code, resp.text)
    resp.raise_for_status()

# 3. 合并分片
def complete_upload(upload_id):
    url = f'{BASE_URL}/upload/{upload_id}/complete'
    resp = requests.post(url)
    print('Complete:', resp.status_code, resp.text)
    resp.raise_for_status()

# 4. 查询状态（可选）
def check_status(upload_id):
    url = f'{BASE_URL}/upload/{upload_id}/status'
    resp = requests.get(url)
    print('Status:', resp.status_code, resp.text)
    resp.raise_for_status()

if __name__ == '__main__':
    filename = 'testfile.txt'
    filedata = b'HelloSwiftStreamTesstFile' * 100000  # 2MB左右
    chunk_size = 102400  # 100KB per chunk
    filetype = 'text/plain'
    filesize = len(filedata)
    
    # 计算文件MD5
    file_md5 = calculate_md5(filedata)
    print(f'文件MD5: {file_md5}')

    upload_id, chunk_count = init_upload(filename, filesize, filetype, chunk_size, file_md5)
    
    # 如果是秒传，直接结束
    if chunk_count == 0:
        print('秒传完成，无需上传分片')
        check_status(upload_id)
        exit()

    for i in range(chunk_count):
        chunk = filedata[i*chunk_size:(i+1)*chunk_size]
        upload_chunk(upload_id, i, chunk)

    check_status(upload_id)
    complete_upload(upload_id)
    check_status(upload_id)
