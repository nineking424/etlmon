# etlmon TUI 테스트 결과 리포트

**테스트 일시:** 2026-02-03 11:29 ~ 11:35 (KST)
**테스트 방법:** tmux를 통한 자동화 테스트 + qa-tester 에이전트
**테스트 환경:** macOS Darwin 24.2.0

---

## 요약

| 항목 | 결과 |
|------|------|
| **TUI 시작** | ✅ 성공 |
| **데이터 표시** | ✅ 성공 |
| **키 바인딩** | ❌ 미구현 |

---

## 테스트 환경 설정

### Node 데몬 시작

```bash
$ ./etlmon-node -c /tmp/etlmon/node-test.yaml
time=2026-02-03T11:29:54.192+09:00 level=INFO msg="starting etlmon node" name=test-node
time=2026-02-03T11:29:54.199+09:00 level=INFO msg="disk collector started" interval=5s
time=2026-02-03T11:29:54.199+09:00 level=INFO msg="path scanner started" paths=1
time=2026-02-03T11:29:54.199+09:00 level=INFO msg="starting API server" address=127.0.0.1:8080
```

### tmux 세션 생성

```bash
$ tmux new-session -d -s tui-test -x 100 -y 25
$ tmux send-keys -t tui-test "./etlmon-ui --node http://127.0.0.1:8080" Enter
```

---

## 테스트 결과

### 1. 초기 화면 캡처 (FS View)

```bash
$ tmux capture-pane -t tui-test -p
```

**출력:**

```
Mount Total     Used      Avail    Use%
/     228.27 GB 202.02 GB 26.26 GB 88.5%
```

**분석:**
- ✅ TUI가 정상적으로 시작됨
- ✅ API에서 파일시스템 데이터를 정상적으로 가져옴
- ✅ 데이터가 올바르게 포맷되어 표시됨
- ⚠️ 헤더와 데이터만 표시됨 (프레임/테두리 없음)

---

### 2. 키 입력 테스트

#### 2.1 '2' 키 (Paths View 전환)

```bash
$ tmux send-keys -t tui-test '2'
$ sleep 2
$ tmux capture-pane -t tui-test -p
```

**출력:**

```
Mount Total     Used      Avail    Use%
/     228.27 GB 202.02 GB 26.26 GB 88.5%
```

**결과:** ❌ **뷰 전환 안됨** - FS View 그대로 유지

---

#### 2.2 'T' 키 (테이블 포맷 토글)

```bash
$ tmux send-keys -t tui-test 'T'
$ sleep 2
$ tmux capture-pane -t tui-test -p
```

**출력:**

```
Mount Total     Used      Avail    Use%
/     228.27 GB 202.02 GB 26.26 GB 88.5%
```

**결과:** ❌ **포맷 변경 안됨**

---

#### 2.3 'q' 키 (종료)

```bash
$ tmux send-keys -t tui-test 'q'
$ sleep 2
$ pgrep -f "etlmon-ui"
62868
```

**결과:** ❌ **종료 안됨** - 프로세스가 계속 실행 중

---

## 코드 분석

### 원인 분석

`ui/views/fs.go` 및 `ui/app.go` 코드 분석 결과:

1. **키 바인딩 미구현**
   - `SetInputCapture()` 호출이 없음
   - 전역 키 핸들러 없음
   - 뷰 전환 로직은 존재하나 (`SwitchView()`) 호출되지 않음

2. **현재 구현된 기능**
   - ✅ API 클라이언트 연결
   - ✅ 파일시스템 데이터 조회
   - ✅ 테이블 렌더링
   - ✅ 뷰 등록 시스템

3. **미구현 기능**
   - ❌ 키 바인딩 (1, 2, T, r, s, q)
   - ❌ 도움말 표시 (?)
   - ❌ 자동 새로고침 연동

### 해당 코드 위치

**ui/app.go:53** - `Run()` 함수:
```go
func (a *App) Run() error {
    // ...
    // SetInputCapture가 호출되지 않음
    return a.tview.Run()
}
```

**필요한 구현:**
```go
a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    switch event.Rune() {
    case '1':
        a.SwitchView("fs")
        return nil
    case '2':
        a.SwitchView("paths")
        return nil
    case 'q':
        a.tview.Stop()
        return nil
    }
    return event
})
```

---

## qa-tester 에이전트 결과

### 에이전트 출력

```
## Test Results Summary

### Test Execution Status: PARTIAL FAILURE

**Step 1: Initial Screen (FS View)**
Mount Total     Used      Avail    Use%
/     228.27 GB 201.94 GB 26.33 GB 88.5%
✓ Application started and displayed filesystem data

**Step 2-6: Key Navigation Tests**
⚠️ All key commands (2, T, 1, r, q) had no effect
❌ Application did NOT exit on 'q' command

### Issues Identified
1. View Switching Not Working
2. Table Toggle Not Working
3. Quit Command Not Working
```

---

## 결론

### 정상 동작하는 기능

| 기능 | 상태 |
|------|------|
| TUI 애플리케이션 시작 | ✅ |
| Node 연결 | ✅ |
| API 데이터 조회 | ✅ |
| 파일시스템 정보 표시 | ✅ |
| 테이블 렌더링 | ✅ |

### 미구현 기능

| 기능 | 예상 키 | 상태 |
|------|---------|------|
| FS 뷰 전환 | `1` | ❌ 미구현 |
| Paths 뷰 전환 | `2` | ❌ 미구현 |
| 테이블 포맷 토글 | `T` | ❌ 미구현 |
| 강제 새로고침 | `r` | ❌ 미구현 |
| 수동 스캔 | `s` | ❌ 미구현 |
| 종료 | `q` | ❌ 미구현 |
| 도움말 | `?` | ❌ 미구현 |

### 권장 조치

1. **긴급:** `ui/app.go`에 `SetInputCapture()` 추가하여 키 바인딩 구현
2. **중요:** 상태바/도움말 영역 추가
3. **선택:** 자동 새로고침 연동

---

## 테스트 환경 정리

```bash
$ tmux kill-session -t tui-test
$ pkill -f "etlmon-ui"
$ pkill -f "etlmon-node"
$ rm -rf /tmp/etlmon
```

---

**문서 버전:** 1.0
**최종 수정일:** 2026-02-03
**작성자:** Claude Code (qa-tester + ralph-loop)
