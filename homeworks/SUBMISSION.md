# Homework Submission

**Họ tên:** Phạm Quang Minh

## Các bài đã hoàn thành

- [x] Bài 1: Statistics APIs
### 1.1 Get Assets Statistics
![alt text](image-3.png)

Có 4 assets

![alt text](image-4.png)

### 1.2 Count Assets by Filter
count
![alt text](image-5.png)

count có type
![alt text](image-6.png)

count có type và status
![alt text](image-7.png)
- [x] Bài 2: Batch Create
tạo được 2 thằng một lúc
![alt text](image-9.png)

![alt text](image-10.png)

invalid domain
![alt text](image-11.png)

limit 100 request
![alt text](image-8.png)
- [x] Bài 3: Batch Delete
giả sử muốn xóa 3 thằng này
![alt text](image-12.png)

![alt text](image-13.png)

check lại -> mất
![alt text](image-15.png)

và xóa lại 3 thằng bên trên -> not found
![alt text](image-14.png)

- [x] Bài 4: Connection Retry

nếu db chưa bật
![alt text](image-16.png)

trong quá trình retry thì bật db lên
![alt text](image-17.png)

- [x] Bài 5: Health Check
db up
![alt text](image-19.png)

db down
![alt text](image-18.png)

Vì db.Stats() lấy thông tin từ connection pool của sql.DB trong app, không phải query trực tiếp xuống DB. Còn db.Ping() mới là cái check DB có thật sự reachable không.

Lấy stats := db.Stats(), rồi Ping() để quyết định 200 hay 503
- [ ] Bài 6: Pagination (Bonus)
- [x] Bài 7: Search (Bonus)

Nếu tồn tại
![alt text](image-20.png)

Nếu không tồn tại 
![alt text](image-21.png)