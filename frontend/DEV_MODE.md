# Frontend Dev Mode - Fast Development

## 🚀 Quick Start (Dev Mode)

```bash
cd frontend
npm run dev
```

**Browser:** http://localhost:5173

## ✅ Advantages

- ⚡ **Instant hot reload** - Thay đổi code → reload ngay lập tức
- 🔍 **Console errors** - Lỗi JS/TS hiện ngay trong browser DevTools
- 🎨 **UI development** - Phát triển UI không cần backend
- 📝 **TypeScript checking** - Lỗi type hiện realtime

## ⚠️ Limitations

- ❌ **No Wails bindings** - API calls sẽ fail (GetScanResults, Scan, etc.)
- ❌ **No events** - Wails Events.On() không hoạt động
- ❌ **No native features** - File dialogs, system integration không có

## 🔧 Dev Workflow

### Phase 1: UI Development (Frontend only)
```bash
cd frontend
npm run dev
```
- Develop UI components
- Style with Tailwind
- Test interactions
- Fix TypeScript errors

### Phase 2: Integration Testing (Full Wails)
```bash
# From project root
npm run build --prefix frontend
./run-gui.sh
```
- Test API integrations
- Test Wails events
- Test full app flow

## 💡 Tips

1. **Mock data** - Create sample data trong components để test UI
2. **Error detection** - Mở Browser DevTools (F12) → Console tab
3. **Hot reload** - Vite tự động reload, không cần restart

## 🐛 Common Dev Mode Errors

### Error: "Events is not defined"
**Cause:** Wails runtime chưa load (bình thường trong dev mode)
**Fix:** Ignore trong dev, hoặc thêm mock

### Error: "GetScanResults is not a function"
**Cause:** Wails bindings không có trong dev mode
**Fix:** Thêm sample data để test UI

## 📊 Example: Mock Data for Testing

```typescript
// In your component
const mockResults = [
  { path: "/Users/test/DerivedData", type: "xcode", size: 1000000000, fileCount: 100, name: "Test Data" }
];

// Use in dev
const results = isDev ? mockResults : await GetScanResults();
```
