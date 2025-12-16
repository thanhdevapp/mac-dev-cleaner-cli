# Mac Dev Cleaner - Yêu Cầu Dự Án

> **Ngày tạo:** 2025-12-15  
> **Stakeholder:** @thanhngo  
> **Status:** Draft - Đang đánh giá

---

## 📋 Tổng Quan

Phát triển một công cụ giúp developers dọn dẹp các thư mục phát triển, giải phóng dung lượng ổ đĩa trên máy Mac.

---

## 🎯 Mục Tiêu Chính

1. **Clean thư mục iOS/Xcode development**
   - `DerivedData`
   - Xcode caches
   - Archives không cần thiết

2. **Clean thư mục Android development**
   - `build/` folders
   - `.gradle/` caches
   - Android SDK caches

3. **Clean cache chung**
   - System caches
   - Application caches

4. **Clean package manager artifacts**
   - `node_modules/`
   - Có thể mở rộng: `Pods/`, `.cargo/`, etc.

---

## 🖥️ Yêu Cầu Giao Diện

| Loại    | Mô tả                                 | Ưu tiên     |
| ------- | ------------------------------------- | ----------- |
| **CLI** | Command line interface cơ bản         | P0 (MVP)    |
| **TUI** | Terminal UI với interactive selection | P1          |
| **GUI** | Desktop app (nếu cần)                 | P2 (Future) |

---

## 🌍 Platform Support

| Platform    | Hỗ trợ    | Ghi chú           |
| ----------- | --------- | ----------------- |
| **macOS**   | ✅ Primary | Target chính      |
| **Windows** | ❓ TBD     | Cần đánh giá thêm |
| **Linux**   | ❓ TBD     | Cần đánh giá thêm |

### Câu hỏi cần quyết định:
- [ ] Chỉ hỗ trợ macOS hay cross-platform?
- [ ] Nếu cross-platform, paths sẽ khác nhau cho mỗi OS

---

## 📦 Đóng Gói & Phân Phối

### Yêu cầu:
- User không cần cài đặt runtime (Node.js, Go, Rust...)
- Dễ dàng cài đặt qua Homebrew (cho macOS)
- Có thể download binary trực tiếp

### Options đã research:

| Stack                 | Pros                         | Cons                            |
| --------------------- | ---------------------------- | ------------------------------- |
| **Go + GoReleaser**   | Fast, simple, cross-platform | Learning curve nếu chưa biết Go |
| **Rust + cargo-dist** | Best performance             | Steeper learning curve          |
| **Bun + compile**     | TypeScript familiar          | Larger binary size              |

> 📄 Chi tiết: Xem [RESEARCH-CLI-DISTRIBUTION.md](./RESEARCH-CLI-DISTRIBUTION.md)

---

## ✅ Acceptance Criteria (MVP)

### Must Have (P0):
- [ ] Scan và liệt kê các thư mục có thể clean
- [ ] Hiển thị size của mỗi thư mục (human-readable)
- [ ] Cho phép chọn thư mục cần xóa
- [ ] Xác nhận trước khi xóa
- [ ] Dry-run mode (preview không xóa thật)

### Should Have (P1):
- [ ] Interactive TUI với arrow key navigation
- [ ] Progress bar khi scanning/deleting
- [ ] Config file để customize paths
- [ ] Presets: `--ios`, `--android`, `--node`, `--all`

### Nice to Have (P2):
- [ ] Auto-detect project types
- [ ] Exclude patterns (whitelist)
- [ ] Report/Summary export
- [ ] Scheduled cleaning

---

## 📁 Thư Mục Target (macOS)

### iOS/Xcode
```
~/Library/Developer/Xcode/DerivedData/
~/Library/Developer/Xcode/Archives/
~/Library/Caches/com.apple.dt.Xcode/
```

### Android
```
~/.gradle/caches/
~/.gradle/wrapper/
~/.android/cache/
*/build/           (trong Android projects)
*/.gradle/         (trong Android projects)
```

### Node.js
```
*/node_modules/
~/.npm/
~/.pnpm-store/
~/.yarn/cache/
```

### General Caches
```
~/Library/Caches/
~/.cache/
```

---

## 🔒 Constraints & Risks

### Safety Requirements:
- ⚠️ **KHÔNG được xóa** nếu user chưa confirm
- ⚠️ **KHÔNG được** xóa system directories
- ⚠️ Phải có **dry-run** mode mặc định
- ⚠️ Log tất cả actions để recover nếu cần

### Technical Constraints:
- Binary size < 20MB (lý tưởng < 10MB)
- Scan performance: < 5s cho ~100 projects
- Memory usage: < 100MB

---

## 📊 Đánh Giá & Phản Hồi

### Các câu hỏi cần feedback:

1. **Platform scope:** Chỉ macOS hay cần Windows/Linux?
2. **UI preference:** TUI đủ hay cần Desktop GUI?
3. **Tech stack:** Go vs Rust vs TypeScript?
4. **Additional folders:** Còn thư mục nào cần clean?
5. **Distribution:** Homebrew đủ hay cần kênh khác?

---

### Phần dành cho reviewer:

**Reviewer:** _______________  
**Ngày review:** _______________

| Mục                 | Approve | Cần sửa | Ghi chú |
| ------------------- | ------- | ------- | ------- |
| Mục tiêu chính      | ☐       | ☐       |         |
| Platform support    | ☐       | ☐       |         |
| MVP features        | ☐       | ☐       |         |
| Tech stack          | ☐       | ☐       |         |
| Safety requirements | ☐       | ☐       |         |

**Nhận xét chung:**

```
[Ghi nhận xét tại đây]
```

---

## 📎 Tài Liệu Liên Quan

- [RESEARCH-CLI-DISTRIBUTION.md](./RESEARCH-CLI-DISTRIBUTION.md) - Research về đóng gói & phân phối
