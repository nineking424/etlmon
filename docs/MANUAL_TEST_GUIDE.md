# etlmon 수동 테스트 가이드

이 문서는 etlmon의 전체 기능을 수동으로 테스트하는 방법을 설명합니다.

## 목차

- [사전 요구사항](#사전-요구사항)
- [빠른 시작](#빠른-시작)
- [환경 설정 및 실행](#환경-설정-및-실행)
- [API 엔드포인트 테스트](#api-엔드포인트-테스트)
- [TUI 화면 레이아웃](#tui-화면-레이아웃)
- [테스트 시나리오](#테스트-시나리오)
  - [카테고리 1: 환경 설정 및 실행](#카테고리-1-환경-설정-및-실행)
  - [카테고리 2: API 엔드포인트 테스트](#카테고리-2-api-엔드포인트-테스트)
  - [카테고리 3: 기본 네비게이션](#카테고리-3-기본-네비게이션)
  - [카테고리 4: 데이터 표시](#카테고리-4-데이터-표시)
  - [카테고리 5: Settings 뷰](#카테고리-5-settings-뷰)
  - [카테고리 6: 오류 처리](#카테고리-6-오류-처리)
  - [카테고리 7: 엣지 케이스](#카테고리-7-엣지-케이스)
  - [카테고리 8: 회귀 테스트](#카테고리-8-회귀-테스트)
  - [카테고리 9: 성능](#카테고리-9-성능)
- [문제 해결](#문제-해결)
- [부록: 빠른 테스트 스크립트](#부록-빠른-테스트-스크립트)

---

## 사전 요구사항

- Go 1.24.0 이상
- curl (API 테스트용)
- jq (JSON 출력 포맷팅용, 선택사항)
- 터미널 에뮬레이터 (최소 80x24 크기 권장)

---

## 빠른 시작

급하게 테스트하려면 아래 명령어만 실행하세요:

```bash
# 1. 빌드
go build -o etlmon-node ./cmd/node
go build -o etlmon-ui ./cmd/ui

# 2. 테스트 환경 설정 (아래 스크립트 참조)
bash quick-test-setup.sh

# 3. Node 실행 (터미널 1)
./etlmon-node -c /tmp/etlmon/node-test.yaml

# 4. UI 실행 (터미널 2)
./etlmon-ui -c /tmp/etlmon/ui-test.yaml
```

---

## 환경 설정 및 실행

### 1단계: 바이너리 빌드

프로젝트 루트 디렉토리에서 실행:

```bash
cd /path/to/etlmon

# Node 데몬 빌드
go build -o etlmon-node ./cmd/node

# TUI 클라이언트 빌드
go build -o etlmon-ui ./cmd/ui
```

**예상 결과:**
- `etlmon-node` 바이너리 생성 (~12MB)
- `etlmon-ui` 바이너리 생성 (~10MB)

**PASS 조건:**
- 두 바이너리가 오류 없이 생성됨
- `./etlmon-node --version`, `./etlmon-ui --version` 실행 가능

**FAIL 조건:**
- 컴파일 오류 발생
- 바이너리 파일이 생성되지 않음

---

### 2단계: 테스트 환경 설정

#### 2.1 테스트 디렉토리 생성

```bash
mkdir -p /tmp/etlmon/data/subdir1/subdir2
mkdir -p /tmp/etlmon/data/logs
```

#### 2.2 Node 설정 파일 생성

`/tmp/etlmon/node-test.yaml` 파일 생성:

```yaml
node:
  listen: 127.0.0.1:8080
  node_name: test-node
  db_path: /tmp/etlmon/etlmon.db

refresh:
  disk: 5s
  default_path_scan: 30s

process:
  patterns:
    - "nifi*"
    - "java*"
    - "etlmon*"
  top_n: 10

logs:
  - name: nifi-app
    path: /tmp/etlmon/data/logs/nifi-app.log
    max_lines: 1000
  - name: test-log
    path: /tmp/etlmon/data/logs/test.log
    max_lines: 500

paths:
  - path: /tmp/etlmon/data
    scan_interval: 10s
    max_depth: 3
    timeout: 30s
```

#### 2.3 UI 설정 파일 생성

`/tmp/etlmon/ui-test.yaml` 파일 생성:

```yaml
nodes:
  - name: test-node
    address: http://127.0.0.1:8080

ui:
  refresh_interval: 2s
  default_node: test-node
```

#### 2.4 테스트 로그 파일 생성

```bash
# 로그 파일 생성
cat > /tmp/etlmon/data/logs/nifi-app.log << 'EOF'
2026-02-14 10:00:00,123 INFO [main] org.apache.nifi.NiFi Starting NiFi...
2026-02-14 10:00:01,456 INFO [main] org.apache.nifi.NiFi NiFi started successfully
2026-02-14 10:01:00,789 WARN [Timer-1] org.apache.nifi.engine.FlowEngine High CPU usage detected
2026-02-14 10:02:00,012 ERROR [Worker-1] org.apache.nifi.processor.ProcessException Failed to process file
EOF

cat > /tmp/etlmon/data/logs/test.log << 'EOF'
[INFO] Test application started
[DEBUG] Processing item 1
[DEBUG] Processing item 2
[WARN] Slow operation detected
[ERROR] Connection timeout
EOF
```

#### 2.5 테스트 파일 생성

```bash
# 일반 파일
echo "테스트 파일 1" > /tmp/etlmon/data/file1.txt
echo "테스트 파일 2" > /tmp/etlmon/data/file2.txt
echo "테스트 파일 3" > /tmp/etlmon/data/subdir1/file3.txt
echo "테스트 파일 4" > /tmp/etlmon/data/subdir1/subdir2/file4.txt

# 큰 파일 (10MB)
dd if=/dev/zero of=/tmp/etlmon/data/large_file.bin bs=1M count=10 2>/dev/null

# 빈 디렉토리
mkdir -p /tmp/etlmon/data/empty_dir
```

---

### 3단계: Node 데몬 실행

**터미널 1**에서 실행:

```bash
./etlmon-node -c /tmp/etlmon/node-test.yaml
```

**예상 출력:**
```
etlmon node starting...
Config loaded: test-node
Database initialized: /tmp/etlmon/etlmon.db
Disk collector started (interval: 5s)
Path collector started (interval: 30s)
Process collector started
Log collector started (2 log files)
API server listening on 127.0.0.1:8080
```

**PASS 조건:**
- 모든 collector가 시작됨
- 오류 메시지 없음
- 포트 8080에서 리스닝 중

**FAIL 조건:**
- "address already in use" 오류
- 설정 파일 파싱 실패
- DB 초기화 실패

---

### 4단계: UI 클라이언트 실행

**터미널 2**에서 실행:

#### 옵션 1: 설정 파일 사용

```bash
./etlmon-ui -c /tmp/etlmon/ui-test.yaml
```

#### 옵션 2: 직접 연결 (설정 파일 없이)

```bash
./etlmon-ui --node http://127.0.0.1:8080
```

**예상 결과:**
- TUI 화면이 나타남
- 헤더에 "Node: test-node" 표시
- 자동 새로고침 동작
- 하단 상태바에 "Last: HH:MM:SS" 표시

**PASS 조건:**
- TUI가 정상적으로 렌더링됨
- 헤더에 노드 이름과 상태 아이콘 표시
- Navbar에 모든 뷰 표시
- 자동 새로고침 작동 (2초마다)

**FAIL 조건:**
- 연결 오류 ("connection refused")
- 화면이 깨지거나 렌더링 안 됨
- 키보드 입력 무반응

---

## API 엔드포인트 테스트

**터미널 3**에서 curl로 API 테스트:

### API-01: Health Check

```bash
curl http://127.0.0.1:8080/health
```

**예상 응답:**
```json
{"status":"ok","node_name":"test-node"}
```

**PASS:** 200 OK, JSON에 status와 node_name 포함
**FAIL:** 연결 오류, status != "ok"

---

### API-02: 파일시스템 사용량 조회

```bash
curl http://127.0.0.1:8080/api/v1/fs | jq
```

**예상 응답:**
```json
{
  "filesystems": [
    {
      "mount_point": "/",
      "device": "/dev/disk1s1",
      "fs_type": "apfs",
      "total_bytes": 494384795648,
      "used_bytes": 234567890123,
      "available_bytes": 259816905525,
      "use_percent": 47.4,
      "collected_at": "2026-02-14T10:00:00Z"
    }
  ]
}
```

**PASS:** 200 OK, filesystems 배열 포함, use_percent 0-100 범위
**FAIL:** 빈 배열, use_percent > 100, collected_at 없음

---

### API-03: 모니터링 경로 조회

```bash
curl http://127.0.0.1:8080/api/v1/paths | jq
```

**예상 응답:**
```json
{
  "paths": [
    {
      "path": "/tmp/etlmon/data",
      "file_count": 5,
      "dir_count": 3,
      "total_size": 10485810,
      "last_scan": "2026-02-14T10:00:00Z",
      "scan_duration": "45ms",
      "status": "OK"
    }
  ]
}
```

**PASS:** 200 OK, file_count > 0, status = "OK"
**FAIL:** file_count = 0 (파일 생성했는데), status = "ERROR"

---

### API-04: 수동 스캔 트리거

```bash
curl -X POST http://127.0.0.1:8080/api/v1/paths/scan | jq
```

**예상 응답:**
```json
{"status":"scan_triggered"}
```

**후속 확인:**
```bash
sleep 2
curl http://127.0.0.1:8080/api/v1/paths | jq
```

**PASS:** 200 OK, status = "scan_triggered", 2초 후 경로 조회 시 last_scan 업데이트됨
**FAIL:** 400/500 오류, last_scan 변경 안 됨

---

### API-05: 프로세스 조회

```bash
curl http://127.0.0.1:8080/api/v1/processes | jq
```

**예상 응답:**
```json
{
  "processes": [
    {
      "pid": 12345,
      "name": "etlmon-node",
      "user": "nineking",
      "cpu_percent": 0.5,
      "memory_rss": 12582912,
      "status": "running",
      "started_at": "2026-02-14T09:00:00Z"
    }
  ]
}
```

**PASS:** 200 OK, processes 배열에 최소 1개 항목 (etlmon-node), cpu_percent >= 0
**FAIL:** 빈 배열, cpu_percent < 0, status 값이 이상함

---

### API-06: 로그 조회

```bash
curl http://127.0.0.1:8080/api/v1/logs | jq
```

**예상 응답:**
```json
{
  "logs": [
    {
      "name": "nifi-app",
      "path": "/tmp/etlmon/data/logs/nifi-app.log",
      "lines": [
        "2026-02-14 10:00:00,123 INFO [main] org.apache.nifi.NiFi Starting NiFi...",
        "2026-02-14 10:00:01,456 INFO [main] org.apache.nifi.NiFi NiFi started successfully"
      ],
      "line_count": 4,
      "collected_at": "2026-02-14T10:05:00Z"
    }
  ]
}
```

**PASS:** 200 OK, logs 배열에 2개 항목 (nifi-app, test-log), lines 배열 포함
**FAIL:** 빈 배열, lines 없음, path 불일치

---

### API-07: Config 조회

```bash
curl http://127.0.0.1:8080/api/v1/config | jq
```

**예상 응답:**
```json
{
  "node": {
    "listen": "127.0.0.1:8080",
    "node_name": "test-node",
    "db_path": "/tmp/etlmon/etlmon.db"
  },
  "refresh": {
    "disk": "5s",
    "default_path_scan": "30s"
  },
  "process": {
    "patterns": ["nifi*", "java*", "etlmon*"],
    "top_n": 10
  },
  "logs": [
    {
      "name": "nifi-app",
      "path": "/tmp/etlmon/data/logs/nifi-app.log",
      "max_lines": 1000
    }
  ],
  "paths": [
    {
      "path": "/tmp/etlmon/data",
      "scan_interval": "10s",
      "max_depth": 3,
      "timeout": "30s"
    }
  ]
}
```

**PASS:** 200 OK, 모든 설정 섹션 포함, node_name = "test-node"
**FAIL:** 일부 섹션 누락, 값 불일치

---

### API-08: Config 수정

```bash
curl -X PUT http://127.0.0.1:8080/api/v1/config \
  -H "Content-Type: application/json" \
  -d '{
    "node": {
      "listen": "127.0.0.1:8080",
      "node_name": "test-node",
      "db_path": "/tmp/etlmon/etlmon.db"
    },
    "refresh": {
      "disk": "5s",
      "default_path_scan": "30s"
    },
    "process": {
      "patterns": ["nifi*", "java*", "python*"],
      "top_n": 15
    },
    "logs": [
      {
        "name": "nifi-app",
        "path": "/tmp/etlmon/data/logs/nifi-app.log",
        "max_lines": 2000
      }
    ],
    "paths": [
      {
        "path": "/tmp/etlmon/data",
        "scan_interval": "10s",
        "max_depth": 3,
        "timeout": "30s"
      }
    ]
  }' | jq
```

**예상 응답:**
```json
{"status":"ok","message":"config updated, restart node to apply changes"}
```

**후속 확인:**
```bash
curl http://127.0.0.1:8080/api/v1/config | jq '.process.top_n'
```

**예상:** `15`

**PASS:** 200 OK, message 포함, 다시 조회 시 변경 사항 반영됨
**FAIL:** 400/500 오류, 설정 변경 안 됨

---

## TUI 화면 레이아웃

### Overview 뷰 (키: 0, 기본)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Filesystem Usage ─────────────────────────────────────────────┐
│ Mount           Usage                          Used     Total  │
│ /               ████████░░░░░░░░░░░░░░░░░░░░  234G    494G   │
│ /System/Vol...  █░░░░░░░░░░░░░░░░░░░░░░░░░░   12G    494G   │
└────────────────────────────────────────────────────────────────┘
┌─ Path Statistics ──────────────────────────────────────────────┐
│ Path                      Files   Dirs  Duration  Status       │
│ /tmp/etlmon/data          5       3     45ms      OK           │
└────────────────────────────────────────────────────────────────┘

View: Overview          Last: 10:05:23          Ready
```

**특징:**
- 상단 1/3: Filesystem Usage (읽기 전용)
- 하단 2/3: Path Statistics (읽기 전용)
- 게이지 바: 30자, 색상 (<75% 녹색, 75-90% 노랑, >90% 빨강)
- 상태: OK=녹색, SCANNING=노랑, ERROR=빨강

---

### Filesystem 뷰 (키: 1)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Filesystem Details ───────────────────────────────────────────┐
│ Mount           Total   Used    Avail   Use%   Usage          │
│ /               494G    234G    260G    47.4%  ██████░░░░     │
│ /System/Vol...  494G    12G     482G    2.4%   █░░░░░░░░░     │
└────────────────────────────────────────────────────────────────┘

View: Filesystem        Last: 10:05:25          Ready
```

**특징:**
- 테이블, 행 선택 가능 (j/k 키)
- 게이지 바: 25자
- Use% 색상 코딩 (<75% 기본, 75-90% 노랑, >90% 빨강)

---

### Paths 뷰 (키: 2)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Path Monitoring ──────────────────────────────────────────────┐
│ Path                      Files   Dirs  Duration  Status       │
│ /tmp/etlmon/data          5       3     45ms      OK           │
└────────────────────────────────────────────────────────────────┘

View: Paths             Last: 10:05:27          Press 's' to scan
```

**특징:**
- 테이블, 행 선택 가능
- `s` 키: 수동 스캔 트리거 (POST /api/v1/paths/scan)
- 상태바에 스캔 안내 메시지

---

### Process 뷰 (키: 3)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Process Monitoring ───────────────────────────────────────────┐
│ PID     User      CPU%    Memory      Status    Elapsed   Name │
│ 12345   nineking  0.5%    12.0MB      running   1h23m     etl..│
│ 12346   nineking  85.3%   512.0MB     running   45m       java │
│ 12347   root      0.1%    8.5MB       sleeping  2d5h      sshd │
└────────────────────────────────────────────────────────────────┘

View: Process           Last: 10:05:29          10 processes
```

**특징:**
- 테이블, 행 선택 가능
- CPU% 색상: >80% 빨강, >50% 노랑
- Status 색상: running=녹색, zombie=빨강, stopped=노랑
- API: GET /api/v1/processes

---

### Logs 뷰 (키: 4)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Logs ─────────────────────────────────────────────────────────┐
│ 10:00:00 nifi-app 2026-02-14 10:00:00,123 INFO [main]...      │
│ 10:00:01 nifi-app 2026-02-14 10:00:01,456 INFO [main]...      │
│ 10:01:00 nifi-app 2026-02-14 10:01:00,789 WARN [Timer-1]...   │
│ 10:02:00 nifi-app 2026-02-14 10:02:00,012 ERROR [Worker-1]... │
│ 10:03:00 test-log [INFO] Test application started             │
│ 10:03:01 test-log [DEBUG] Processing item 1                   │
└────────────────────────────────────────────────────────────────┘

View: Logs              Last: 10:05:31          Scrollable
```

**특징:**
- TextView, 스크롤 가능 (j/k, PageUp/PageDown)
- 포맷: `[teal]HH:MM:SS[-] [aqua]log-name[-] log line`
- 자동으로 하단 스크롤 (최신 로그)
- 빈 상태: "(No log entries)" 표시
- API: GET /api/v1/logs

---

### Settings 뷰 (키: 5)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─────────────┬──────────────────────────────────────────────────┐
│ Process     │ ┌─ Process Patterns ────────────────────────────┐│
│ Logs        │ │ Pattern                                      │││
│ Paths       │ │ nifi*                                        │││
│             │ │ java*                                        │││
│             │ │ etlmon*                                      │││
│             │ └──────────────────────────────────────────────┘││
│             │ Top N Display: [10              ]              ││
│             │                                                 ││
│             │ [a] Add  [d] Delete  [s] Save  [Tab] Switch    ││
└─────────────┴──────────────────────────────────────────────────┘

View: Settings          Last: 10:05:33          Tab to switch focus
```

**특징:**
- 좌측 사이드바 (섹션 선택) + 우측 컨텐츠
- Tab/Shift+Tab: 사이드바 ↔ 컨텐츠 전환
- `a`: 항목 추가 (모달 폼)
- `d`: 항목 삭제 (선택된 행)
- `s`: 설정 저장 (PUT /api/v1/config)
- 섹션:
  - Process: 패턴 리스트 + Top N 입력 필드
  - Logs: Name | Path | MaxLines 테이블
  - Paths: Path | Interval (포맷: 60s, 5m, 1h) | MaxDepth 테이블
- 모달 폼: Esc/Cancel로 닫기
- 모달 열려있을 때 전역 키 블록됨
- Dirty 상태일 때 자동 새로고침 건너뜀

---

### Help 뷰 (키: ? 또는 h)

```
┌─ ETLMON ──────────────────────────────────────────────────────┐
│ Node: test-node ● OK                                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
<0> Overview  <1> FS  <2> Paths  <3> Process  <4> Logs  <5> Settings  │  ?=help  r=refresh  s=scan  q=quit

┌─ Keyboard Shortcuts ───────────────────────────────────────────┐
│                                                                 │
│ Global Keys:                                                   │
│   0-5         Switch between views                             │
│   ? / h       Show this help                                   │
│   r           Refresh current view                             │
│   q           Quit application                                 │
│   Ctrl+C      Force quit                                       │
│                                                                 │
│ Paths View:                                                    │
│   s           Trigger manual scan                              │
│                                                                 │
│ Settings View:                                                 │
│   Tab         Switch focus (sidebar ↔ content)                │
│   a           Add new entry                                    │
│   d           Delete selected entry                            │
│   s           Save configuration                               │
│                                                                 │
│ Press any key to close...                                      │
└─────────────────────────────────────────────────────────────────┘

View: Help              Last: 10:05:35
```

**특징:**
- 모달 TextView
- 아무 키나 눌러서 닫기
- 이전 뷰로 복귀

---

## 테스트 시나리오

### 카테고리 1: 환경 설정 및 실행

#### ENV-01: 빌드 및 바이너리 생성

**단계:**
1. `go build -o etlmon-node ./cmd/node` 실행
2. `go build -o etlmon-ui ./cmd/ui` 실행
3. `ls -lh etlmon-*` 확인

**예상 결과:**
- 두 바이너리 파일 생성
- 파일 크기 ~10-15MB

**PASS:** 두 파일 모두 생성됨, 실행 권한 있음
**FAIL:** 컴파일 오류, 파일 없음

---

#### ENV-02: 테스트 환경 설정

**단계:**
1. 부록의 `quick-test-setup.sh` 스크립트 실행
2. `/tmp/etlmon/` 디렉토리 확인
3. 설정 파일 2개 확인 (node-test.yaml, ui-test.yaml)

**예상 결과:**
- `/tmp/etlmon/data/` 디렉토리 존재
- 설정 파일 2개 생성됨
- 테스트 파일 5개 생성됨

**PASS:** 모든 파일 및 디렉토리 존재
**FAIL:** 디렉토리/파일 누락

---

#### ENV-03: Node 데몬 실행 및 확인

**단계:**
1. `./etlmon-node -c /tmp/etlmon/node-test.yaml` 실행
2. 출력 메시지 확인
3. 다른 터미널에서 `curl http://127.0.0.1:8080/health` 실행

**예상 결과:**
```
etlmon node starting...
Config loaded: test-node
Database initialized: /tmp/etlmon/etlmon.db
Disk collector started (interval: 5s)
Path collector started (interval: 30s)
Process collector started
Log collector started (2 log files)
API server listening on 127.0.0.1:8080
```

**PASS:** 모든 collector 시작, health 체크 성공
**FAIL:** 오류 메시지, health 체크 실패

---

#### ENV-04: UI 클라이언트 실행

**단계:**
1. **옵션 A:** `./etlmon-ui -c /tmp/etlmon/ui-test.yaml` 실행
2. **옵션 B:** `./etlmon-ui --node http://127.0.0.1:8080` 실행
3. TUI 화면 확인

**예상 결과:**
- TUI 화면 정상 렌더링
- 헤더에 "Node: test-node" 표시
- Overview 뷰 (기본 뷰) 표시

**PASS:** TUI 화면 나타남, 헤더 정상, 데이터 로드됨
**FAIL:** 연결 오류, 화면 깨짐, 무한 로딩

---

### 카테고리 2: API 엔드포인트 테스트

> API-01 ~ API-08은 위 "API 엔드포인트 테스트" 섹션 참조

---

### 카테고리 3: 기본 네비게이션

#### TC-001: 뷰 전환 (0-5 키)

**단계:**
1. UI 실행
2. `0` ~ `5` 키 순서대로 누름
3. 각 뷰가 표시되는지 확인

**예상 결과:**
- `0`: Overview 뷰
- `1`: Filesystem 뷰
- `2`: Paths 뷰
- `3`: Process 뷰
- `4`: Logs 뷰
- `5`: Settings 뷰

**PASS:** 모든 뷰가 정상적으로 표시됨, Navbar에서 활성 뷰 하이라이트됨
**FAIL:** 뷰 전환 안 됨, 화면 깨짐

---

#### TC-002: Help 뷰 열기/닫기

**단계:**
1. Overview 뷰에서 `?` 또는 `h` 키 입력
2. Help 뷰 내용 확인
3. 아무 키나 입력 (예: Enter)
4. 이전 뷰(Overview)로 돌아가는지 확인

**예상 결과:**
- Help 모달이 나타남
- 키보드 단축키 목록 표시
- 아무 키나 누르면 이전 뷰로 복귀

**PASS:** Help 열림, 내용 표시, 닫힘 후 이전 뷰 복귀
**FAIL:** Help 안 열림, 닫혀도 이전 뷰 안 나옴

---

#### TC-003: 자동 새로고침

**단계:**
1. UI 실행 (refresh_interval: 2s)
2. Overview 뷰에서 상태바 "Last: HH:MM:SS" 확인
3. 2초 기다림
4. "Last" 시간이 업데이트되는지 확인

**예상 결과:**
- 2초마다 자동 새로고침
- 상태바 "Last" 시간 업데이트
- 데이터 갱신

**PASS:** 2초마다 "Last" 시간 변경, 데이터 새로고침됨
**FAIL:** 자동 새로고침 안 됨, 시간 고정

---

#### TC-004: 수동 새로고침 (r)

**단계:**
1. Overview 뷰에서 `r` 키 입력
2. 상태바 "Last" 시간 확인
3. 즉시 업데이트되는지 확인

**예상 결과:**
- `r` 키 입력 즉시 새로고침
- "Last" 시간 즉시 업데이트
- 상태바에 "Refreshing..." 잠깐 표시 (선택사항)

**PASS:** 즉시 새로고침, 데이터 갱신
**FAIL:** `r` 키 무반응, 새로고침 안 됨

---

### 카테고리 4: 데이터 표시

#### TC-005: Overview FS 데이터

**단계:**
1. Overview 뷰로 이동 (키 `0`)
2. 상단 "Filesystem Usage" 섹션 확인
3. Mount, Usage 게이지, Used, Total 값 확인

**예상 결과:**
- 최소 1개 파일시스템 (루트 `/`) 표시
- Usage 게이지 30자, 색상 코딩
- Used/Total 값 표시 (예: 234G / 494G)

**PASS:** FS 데이터 표시, 게이지 렌더링, 값 정확
**FAIL:** 빈 화면, 게이지 없음, 값 0

---

#### TC-006: Overview Path 데이터

**단계:**
1. Overview 뷰에서 하단 "Path Statistics" 섹션 확인
2. `/tmp/etlmon/data` 경로 확인
3. Files, Dirs, Duration, Status 값 확인

**예상 결과:**
- `/tmp/etlmon/data` 경로 표시
- Files: 5, Dirs: 3 (또는 생성한 파일 수)
- Status: "OK" (녹색)
- Duration: ~ms 단위

**PASS:** Path 데이터 표시, 값 정확, 상태 색상 녹색
**FAIL:** 빈 화면, Files=0, Status=ERROR

---

#### TC-007: Overview 부분 실패

**단계:**
1. Node 데몬 중지
2. UI는 계속 실행
3. Overview 뷰에서 FS/Path 섹션 확인

**예상 결과:**
- FS API 실패 시 "Error fetching filesystem data" 표시
- Path API 실패 시 "Error fetching path data" 표시
- 한쪽 실패해도 다른 쪽은 표시 (부분 실패 처리)

**PASS:** 오류 메시지 표시, 앱 크래시 안 함
**FAIL:** 앱 종료, 무한 로딩, 화면 깨짐

---

#### TC-008: FS 뷰 전체 데이터

**단계:**
1. FS 뷰로 이동 (키 `1`)
2. 테이블 확인 (Mount, Total, Used, Avail, Use%, Usage 게이지)
3. j/k 키로 행 선택 테스트

**예상 결과:**
- 모든 마운트 포인트 표시
- 게이지 바 25자, 색상 코딩
- j/k로 행 선택 가능

**PASS:** 모든 컬럼 표시, 게이지 정상, 행 선택 작동
**FAIL:** 일부 컬럼 누락, 게이지 깨짐, 선택 안 됨

---

#### TC-009: Paths 뷰 스캔 트리거

**단계:**
1. Paths 뷰로 이동 (키 `2`)
2. 테스트 파일 추가: `echo "new file" > /tmp/etlmon/data/newfile.txt`
3. `s` 키 입력 (수동 스캔)
4. 2초 후 Files 카운트 확인

**예상 결과:**
- `s` 키 입력 시 상태바에 "Scan triggered" 메시지
- 2초 후 Files 카운트 증가 (5 → 6)
- Status: "OK"

**PASS:** 스캔 트리거됨, Files 카운트 증가
**FAIL:** `s` 키 무반응, 카운트 변경 안 됨

---

#### TC-010: Paths 빈 상태

**단계:**
1. 설정 파일에서 존재하지 않는 경로 추가 (`/nonexistent/path`)
2. Node 재시작
3. Paths 뷰 확인

**예상 결과:**
- 경로 표시되지만 Files=0, Dirs=0
- Status: "ERROR" (빨강) 또는 "OK"
- 오류 메시지 표시 (선택사항)

**PASS:** 빈 상태 처리, 앱 크래시 안 함
**FAIL:** 앱 종료, 무한 로딩

---

#### TC-011: Process 색상 코딩

**단계:**
1. Process 뷰로 이동 (키 `3`)
2. CPU% 컬럼 확인
3. Status 컬럼 확인

**예상 결과:**
- CPU% > 80%: 빨강
- CPU% > 50%: 노랑
- CPU% < 50%: 기본 색
- Status "running": 녹색
- Status "zombie": 빨강
- Status "stopped": 노랑

**PASS:** 색상 코딩 정확
**FAIL:** 모든 색상 동일, 색상 안 나옴

---

#### TC-012: Logs 스크롤

**단계:**
1. Logs 뷰로 이동 (키 `4`)
2. j/k 키로 스크롤 테스트
3. PageUp/PageDown 테스트

**예상 결과:**
- 로그 라인 표시 (시간 + 로그명 + 내용)
- j/k로 한 줄씩 스크롤
- PageUp/PageDown으로 페이지 단위 스크롤
- 자동으로 하단(최신 로그)에 위치

**PASS:** 스크롤 정상 작동, 로그 포맷 정확
**FAIL:** 스크롤 안 됨, 로그 안 보임

---

#### TC-013: Logs 빈 상태

**단계:**
1. 로그 파일 비우기: `> /tmp/etlmon/data/logs/nifi-app.log`
2. `> /tmp/etlmon/data/logs/test.log`
3. Logs 뷰 확인

**예상 결과:**
- "(No log entries)" 메시지 표시
- 빈 화면 또는 안내 메시지

**PASS:** 빈 상태 메시지 표시
**FAIL:** 오류 발생, 이전 로그 계속 표시

---

### 카테고리 5: Settings 뷰

#### TC-014: 섹션 네비게이션

**단계:**
1. Settings 뷰로 이동 (키 `5`)
2. 좌측 사이드바에서 "Process", "Logs", "Paths" 섹션 선택
3. 우측 컨텐츠 영역 변화 확인

**예상 결과:**
- Process 섹션: 패턴 리스트 + Top N 입력 필드
- Logs 섹션: Name | Path | MaxLines 테이블
- Paths 섹션: Path | Interval | MaxDepth 테이블

**PASS:** 섹션 전환 정상, 컨텐츠 올바르게 표시
**FAIL:** 섹션 전환 안 됨, 컨텐츠 깨짐

---

#### TC-015: Process 패턴 추가

**단계:**
1. Settings → Process 섹션 선택
2. Tab 키로 우측 컨텐츠로 포커스 이동
3. `a` 키 입력 (Add)
4. 모달 폼에서 패턴 입력: `python*`
5. Enter로 저장

**예상 결과:**
- 모달 폼 나타남
- 입력 필드에 "python*" 입력
- Enter 후 모달 닫힘
- 패턴 리스트에 "python*" 추가됨

**PASS:** 모달 열림, 입력 가능, 저장 후 리스트 업데이트
**FAIL:** 모달 안 열림, 입력 안 됨, 저장 안 됨

---

#### TC-016: Process 패턴 삭제

**단계:**
1. Settings → Process 섹션
2. Tab 키로 패턴 리스트에 포커스
3. j/k로 삭제할 패턴 선택
4. `d` 키 입력

**예상 결과:**
- 확인 모달 또는 즉시 삭제
- 패턴 리스트에서 제거됨

**PASS:** 패턴 삭제됨
**FAIL:** `d` 키 무반응, 삭제 안 됨

---

#### TC-017: Log 모니터 추가

**단계:**
1. Settings → Logs 섹션
2. `a` 키 입력
3. 모달 폼에서 입력:
   - Name: `system-log`
   - Path: `/var/log/system.log`
   - MaxLines: `2000`
4. Enter로 저장

**예상 결과:**
- 모달 폼 3개 필드 표시
- 입력 후 저장
- 테이블에 새 로그 추가됨

**PASS:** 모달 열림, 입력 가능, 저장 후 테이블 업데이트
**FAIL:** 모달 안 열림, 입력 안 됨, 저장 안 됨

---

#### TC-018: Path 추가

**단계:**
1. Settings → Paths 섹션
2. `a` 키 입력
3. 모달 폼에서 입력:
   - Path: `/var/log/nifi`
   - Interval: `300s` (또는 `5m`)
   - MaxDepth: `5`
   - Timeout: `60s` (또는 `1m`)
4. Enter로 저장

**예상 결과:**
- 모달 폼 4개 필드 표시
- Interval 포맷 변환 (300s → 5m)
- 테이블에 새 경로 추가됨

**PASS:** 모달 열림, 입력 가능, 저장 후 테이블 업데이트, Interval 포맷 정확
**FAIL:** 모달 안 열림, 입력 안 됨, Interval 포맷 깨짐

---

#### TC-019: 설정 저장

**단계:**
1. Settings에서 항목 추가/삭제 (위 TC-015 ~ TC-018)
2. `s` 키 입력 (Save)
3. 상태바 메시지 확인
4. 다른 터미널에서 `curl http://127.0.0.1:8080/api/v1/config | jq` 확인

**예상 결과:**
- `s` 키 입력 시 상태바에 "Configuration saved, restart node to apply" 메시지
- API 조회 시 변경 사항 반영됨

**PASS:** 저장 메시지 표시, API 응답에 변경 사항 반영
**FAIL:** `s` 키 무반응, API 변경 안 됨

---

#### TC-020: 모달 취소

**단계:**
1. Settings → Process → `a` 키
2. 모달 폼에서 패턴 입력 시작
3. Esc 키 입력

**예상 결과:**
- 모달 닫힘
- 입력 내용 버려짐
- 이전 화면으로 복귀

**PASS:** Esc로 모달 닫힘, 입력 저장 안 됨
**FAIL:** Esc 무반응, 입력 저장됨

---

#### TC-021: 모달 키 보호

**단계:**
1. Settings → Process → `a` 키 (모달 열기)
2. 모달 열린 상태에서 `q` 키 입력
3. 앱이 종료되지 않는지 확인

**예상 결과:**
- 모달 열린 상태에서 전역 키(`q`, `r`, `0-5` 등) 블록됨
- `q` 입력 시 앱 종료 안 됨
- 모달 닫은 후 전역 키 다시 작동

**PASS:** 모달 열린 동안 전역 키 블록, 모달 닫은 후 복구
**FAIL:** 모달 중 `q`로 앱 종료됨

---

#### TC-022: 자동 새로고침 보호

**단계:**
1. Settings → Logs → `a` 키 (모달 열기)
2. 모달에서 Name 입력 중 2초 대기 (auto-refresh interval)
3. 입력 내용이 사라지지 않는지 확인

**예상 결과:**
- 모달 열린 상태에서 자동 새로고침 건너뜀
- 입력 중인 데이터 유지됨

**PASS:** 자동 새로고침 건너뛰어짐, 입력 데이터 유지
**FAIL:** 자동 새로고침으로 입력 데이터 손실

---

#### TC-023: Dirty 플래그 보호

**단계:**
1. Settings → Process → 패턴 추가 (`a`)
2. 저장하지 않고 (`s` 안 누름)
3. 2초 대기 (auto-refresh interval)
4. 추가한 패턴이 사라지지 않는지 확인

**예상 결과:**
- Dirty 상태일 때 자동 새로고침 건너뜀
- 저장하지 않은 변경 사항 유지됨

**PASS:** Dirty 상태에서 자동 새로고침 건너뜀, 데이터 유지
**FAIL:** 자동 새로고침으로 변경 사항 손실

---

### 카테고리 6: 오류 처리

#### TC-024: API 연결 오류

**단계:**
1. UI 실행 중
2. Node 데몬 강제 종료 (Ctrl+C)
3. UI에서 `r` 키 (새로고침)

**예상 결과:**
- 상태바에 오류 메시지 표시 (예: "Connection refused")
- 앱 크래시 안 함
- 이전 데이터 유지 또는 빈 화면

**PASS:** 오류 메시지 표시, 앱 계속 실행
**FAIL:** 앱 종료, 무한 로딩, 화면 깨짐

---

#### TC-025: Settings 저장 오류

**단계:**
1. Settings에서 잘못된 Path 추가 (예: 빈 문자열)
2. `s` 키로 저장 시도

**예상 결과:**
- API 400 오류 또는 검증 오류 메시지
- 상태바에 "Failed to save configuration: ..." 표시
- 설정 변경 안 됨

**PASS:** 오류 메시지 표시, 저장 안 됨
**FAIL:** 잘못된 설정 저장됨, 앱 크래시

---

#### TC-026: Paths 스캔 오류

**단계:**
1. Paths 뷰에서 권한 없는 경로 스캔 (`/root/private`)
2. `s` 키 입력

**예상 결과:**
- 상태바에 "Scan triggered" 또는 "Scan failed" 메시지
- Status 컬럼에 "ERROR" (빨강) 표시
- 앱 크래시 안 함

**PASS:** 오류 상태 표시, 앱 계속 실행
**FAIL:** 앱 종료, 무한 로딩

---

### 카테고리 7: 엣지 케이스

#### TC-027: 빈 응답

**단계:**
1. Node 설정에서 paths, logs, process 모두 비움
2. Node 재시작
3. UI에서 각 뷰 확인

**예상 결과:**
- FS 뷰: 최소 1개 파일시스템 (루트)
- Paths 뷰: 빈 테이블 또는 "No paths configured"
- Process 뷰: 최소 1개 프로세스 (etlmon-node)
- Logs 뷰: "(No log entries)"

**PASS:** 빈 상태 처리, 앱 정상 작동
**FAIL:** 앱 크래시, 화면 깨짐

---

#### TC-028: 대용량 데이터

**단계:**
1. 대량 파일 생성: `for i in {1..1000}; do touch /tmp/etlmon/data/file$i.txt; done`
2. Paths 뷰에서 스캔 (`s`)
3. Files 카운트 확인

**예상 결과:**
- Files: 1000+ 표시
- 스캔 시간 증가 (Duration: ~초 단위)
- UI 응답성 유지

**PASS:** 대용량 처리, UI 반응 유지
**FAIL:** UI 멈춤, 타임아웃, 카운트 오류

---

#### TC-029: 긴 값

**단계:**
1. 매우 긴 경로 추가: `/tmp/etlmon/data/very/long/path/that/exceeds/normal/terminal/width/and/causes/wrapping/issues`
2. Paths 뷰 확인

**예상 결과:**
- 경로 표시 (잘림 또는 줄바꿈)
- "..." 또는 말줄임표 처리
- 테이블 레이아웃 깨지지 않음

**PASS:** 긴 값 처리, 레이아웃 유지
**FAIL:** 테이블 깨짐, 화면 넘침

---

#### TC-030: 빠른 뷰 전환

**단계:**
1. `0` ~ `5` 키를 빠르게 반복 입력 (1초에 10회)
2. 앱 반응성 확인

**예상 결과:**
- 모든 뷰 전환 처리
- UI 멈춤 없음
- 최종적으로 마지막 입력 뷰 표시

**PASS:** 모든 입력 처리, UI 반응 유지
**FAIL:** UI 멈춤, 입력 무시, 크래시

---

#### TC-031: 동시 새로고침

**단계:**
1. 자동 새로고침 활성화 (2초 간격)
2. 정확히 자동 새로고침 시점에 `r` 키 입력
3. 중복 요청 처리 확인

**예상 결과:**
- 중복 요청 방지 또는 무시
- UI 응답성 유지
- 데이터 일관성 유지

**PASS:** 중복 처리, 앱 정상 작동
**FAIL:** 다중 요청으로 UI 멈춤, 데이터 깨짐

---

#### TC-032: 다중 모달

**단계:**
1. Settings → Process → `a` (모달 열기)
2. 모달 열린 상태에서 Help (`?`) 시도
3. 모달이 닫히지 않는지 확인

**예상 결과:**
- 모달 열린 동안 다른 모달/뷰 열리지 않음
- 또는 현재 모달 닫고 Help 열림 (구현에 따라)

**PASS:** 모달 충돌 없음, 동작 명확
**FAIL:** 다중 모달로 화면 깨짐

---

### 카테고리 8: 회귀 테스트

#### TC-033: Settings 모달 포커스 버그

**이슈:** 모달 폼에서 Tab 키 입력 시 사이드바로 포커스 이동하여 편집 불가

**단계:**
1. Settings → Logs → `a` (모달 열기)
2. Name 필드에서 Tab 키 입력
3. Path 필드로 포커스 이동 확인

**예상 결과:**
- Tab 키로 모달 내 필드 간 이동
- 사이드바로 포커스 이동 안 됨

**PASS:** 모달 내 Tab 동작, 사이드바 이동 안 됨
**FAIL:** Tab으로 사이드바 이동, 편집 불가

---

#### TC-034: Settings 자동 새로고침 데이터 손실

**이슈:** 자동 새로고침으로 편집 중인 데이터 사라짐

**단계:**
1. Settings → Paths → `a` (모달 열기)
2. Path 입력 중 2초 대기 (auto-refresh)
3. 입력 내용 유지 확인

**예상 결과:**
- 모달 열린 동안 자동 새로고침 건너뜀
- 입력 데이터 유지

**PASS:** 자동 새로고침 건너뜀, 데이터 유지
**FAIL:** 입력 데이터 손실

---

#### TC-035: Settings 저장 키 전파

**이슈:** Settings 뷰에서 `s` 키 입력 시 Paths 뷰의 스캔 트리거도 실행됨

**단계:**
1. Settings 뷰에서 패턴 추가
2. `s` 키로 저장
3. Paths 스캔이 트리거되지 않는지 확인

**예상 결과:**
- Settings 저장만 실행
- Paths 스캔 트리거 안 됨

**PASS:** `s` 키가 Settings 내에서만 처리
**FAIL:** Paths 스캔도 함께 실행됨

---

### 카테고리 9: 성능

#### TC-036: 자동 새로고침 부하

**단계:**
1. refresh_interval을 1s로 설정
2. 10분간 UI 실행
3. CPU/메모리 사용량 확인

**예상 결과:**
- CPU: <5% (평균)
- 메모리: <50MB 증가
- UI 반응성 유지

**PASS:** 리소스 사용량 안정, UI 반응 유지
**FAIL:** CPU 100%, 메모리 누수, UI 멈춤

**측정 방법:**
```bash
# macOS
top -pid $(pgrep etlmon-ui)

# Linux
htop -p $(pgrep etlmon-ui)
```

---

#### TC-037: 대용량 로그

**단계:**
1. 대용량 로그 생성: `for i in {1..10000}; do echo "Log line $i with some content" >> /tmp/etlmon/data/logs/test.log; done`
2. Logs 뷰로 이동
3. 스크롤 테스트

**예상 결과:**
- max_lines 제한 적용 (1000줄)
- 스크롤 부드러움
- UI 응답성 유지

**PASS:** max_lines 제한 작동, 스크롤 부드러움
**FAIL:** 10000줄 모두 로드, UI 멈춤

---

## 문제 해결

### 문제: Node 시작 실패 - "address already in use"

**원인:** 포트 8080이 이미 사용 중

**해결:**
```bash
# 사용 중인 프로세스 확인
lsof -i :8080

# 다른 포트 사용
# node-test.yaml에서 listen: 127.0.0.1:9090 으로 변경
```

---

### 문제: TUI 연결 실패 - "connection refused"

**원인:** Node 데몬이 실행되지 않음

**해결:**
1. Node 데몬이 실행 중인지 확인
2. 주소와 포트가 올바른지 확인
3. 방화벽 설정 확인

```bash
# Node 프로세스 확인
ps aux | grep etlmon-node

# 포트 리스닝 확인
lsof -i :8080
```

---

### 문제: 경로 스캔 결과가 0

**원인:** 설정된 경로가 존재하지 않거나 권한 없음

**해결:**
```bash
# 경로 존재 확인
ls -la /tmp/etlmon/data

# 권한 확인
stat /tmp/etlmon/data

# 파일 생성 권한 확인
touch /tmp/etlmon/data/test.txt
```

---

### 문제: 데이터베이스 오류

**원인:** DB 파일 손상 또는 권한 문제

**해결:**
```bash
# DB 파일 삭제 후 재시작
rm /tmp/etlmon/etlmon.db*
./etlmon-node -c /tmp/etlmon/node-test.yaml
```

---

### 문제: Settings 저장 안 됨

**원인:** API 권한 오류 또는 잘못된 설정 값

**해결:**
1. Node 로그 확인 (터미널 1)
2. API 직접 테스트:
```bash
curl -X PUT http://127.0.0.1:8080/api/v1/config \
  -H "Content-Type: application/json" \
  -d @/tmp/etlmon/test-config.json
```
3. 설정 파일 YAML 문법 검증

---

### 문제: TUI 화면 깨짐

**원인:** 터미널 크기 너무 작음

**해결:**
- 터미널 크기 최소 80x24로 조정
- `resize` 명령어 실행 (Linux)
- iTerm2/Terminal.app 재시작 (macOS)

---

### 문제: 로그가 표시되지 않음

**원인:** 로그 파일 권한 또는 경로 오류

**해결:**
```bash
# 로그 파일 존재 확인
ls -la /tmp/etlmon/data/logs/

# 로그 파일 읽기 권한 확인
cat /tmp/etlmon/data/logs/nifi-app.log

# Node 설정에서 로그 경로 확인
cat /tmp/etlmon/node-test.yaml | grep -A 10 "logs:"
```

---

### 문제: Process 뷰에 프로세스 없음

**원인:** 패턴 매칭 실패 또는 top_n=0

**해결:**
1. 설정에서 패턴 확인 (`*` 와일드카드 사용)
2. top_n 값 확인 (>0)
3. 패턴에 맞는 프로세스 실행 중인지 확인:
```bash
ps aux | grep nifi
ps aux | grep java
```

---

## 부록: 빠른 테스트 스크립트

아래 스크립트를 사용하면 전체 테스트 환경을 한 번에 설정할 수 있습니다.

### `quick-test-setup.sh`

```bash
#!/bin/bash
# quick-test-setup.sh
# etlmon 테스트 환경 자동 설정 스크립트

set -e

echo "=== etlmon 테스트 환경 설정 시작 ==="

# 1. 테스트 디렉토리 생성
echo "1. 테스트 디렉토리 생성 중..."
mkdir -p /tmp/etlmon/data/subdir1/subdir2
mkdir -p /tmp/etlmon/data/empty_dir
mkdir -p /tmp/etlmon/data/logs

# 2. Node 설정 파일 생성
echo "2. Node 설정 파일 생성 중..."
cat > /tmp/etlmon/node-test.yaml << 'EOF'
node:
  listen: 127.0.0.1:8080
  node_name: test-node
  db_path: /tmp/etlmon/etlmon.db

refresh:
  disk: 5s
  default_path_scan: 30s

process:
  patterns:
    - "nifi*"
    - "java*"
    - "etlmon*"
  top_n: 10

logs:
  - name: nifi-app
    path: /tmp/etlmon/data/logs/nifi-app.log
    max_lines: 1000
  - name: test-log
    path: /tmp/etlmon/data/logs/test.log
    max_lines: 500

paths:
  - path: /tmp/etlmon/data
    scan_interval: 10s
    max_depth: 3
    timeout: 30s
EOF

# 3. UI 설정 파일 생성
echo "3. UI 설정 파일 생성 중..."
cat > /tmp/etlmon/ui-test.yaml << 'EOF'
nodes:
  - name: test-node
    address: http://127.0.0.1:8080

ui:
  refresh_interval: 2s
  default_node: test-node
EOF

# 4. 테스트 파일 생성
echo "4. 테스트 파일 생성 중..."
echo "테스트 파일 1" > /tmp/etlmon/data/file1.txt
echo "테스트 파일 2" > /tmp/etlmon/data/file2.txt
echo "테스트 파일 3" > /tmp/etlmon/data/subdir1/file3.txt
echo "테스트 파일 4" > /tmp/etlmon/data/subdir1/subdir2/file4.txt

# 5. 큰 파일 생성 (10MB)
echo "5. 큰 파일 생성 중 (10MB)..."
dd if=/dev/zero of=/tmp/etlmon/data/large_file.bin bs=1M count=10 2>/dev/null

# 6. 로그 파일 생성
echo "6. 로그 파일 생성 중..."
cat > /tmp/etlmon/data/logs/nifi-app.log << 'EOF'
2026-02-14 10:00:00,123 INFO [main] org.apache.nifi.NiFi Starting NiFi...
2026-02-14 10:00:01,456 INFO [main] org.apache.nifi.NiFi NiFi started successfully
2026-02-14 10:01:00,789 WARN [Timer-1] org.apache.nifi.engine.FlowEngine High CPU usage detected
2026-02-14 10:02:00,012 ERROR [Worker-1] org.apache.nifi.processor.ProcessException Failed to process file
EOF

cat > /tmp/etlmon/data/logs/test.log << 'EOF'
[INFO] Test application started
[DEBUG] Processing item 1
[DEBUG] Processing item 2
[WARN] Slow operation detected
[ERROR] Connection timeout
EOF

# 7. 권한 확인
echo "7. 권한 확인 중..."
chmod -R 755 /tmp/etlmon

echo ""
echo "=== 테스트 환경 설정 완료! ==="
echo ""
echo "디렉토리 구조:"
tree -L 3 /tmp/etlmon 2>/dev/null || find /tmp/etlmon -type f -o -type d | head -20
echo ""
echo "다음 명령어로 실행하세요:"
echo ""
echo "  [터미널 1] Node 데몬 실행:"
echo "    ./etlmon-node -c /tmp/etlmon/node-test.yaml"
echo ""
echo "  [터미널 2] UI 클라이언트 실행 (옵션 A - 설정 파일):"
echo "    ./etlmon-ui -c /tmp/etlmon/ui-test.yaml"
echo ""
echo "  [터미널 2] UI 클라이언트 실행 (옵션 B - 직접 연결):"
echo "    ./etlmon-ui --node http://127.0.0.1:8080"
echo ""
echo "  [터미널 3] API 테스트:"
echo "    curl http://127.0.0.1:8080/health"
echo "    curl http://127.0.0.1:8080/api/v1/fs | jq"
echo "    curl http://127.0.0.1:8080/api/v1/paths | jq"
echo "    curl http://127.0.0.1:8080/api/v1/processes | jq"
echo "    curl http://127.0.0.1:8080/api/v1/logs | jq"
echo "    curl http://127.0.0.1:8080/api/v1/config | jq"
echo ""
echo "정리 명령어:"
echo "  rm -rf /tmp/etlmon"
echo ""
```

### 사용 방법

1. 스크립트 저장:
```bash
cat > quick-test-setup.sh << 'EOF'
[위의 스크립트 내용 붙여넣기]
EOF
```

2. 실행 권한 부여:
```bash
chmod +x quick-test-setup.sh
```

3. 실행:
```bash
./quick-test-setup.sh
```

---

## 테스트 체크리스트

아래 체크리스트를 사용하여 테스트 진행 상황을 추적하세요:

### 환경 설정
- [ ] ENV-01: 바이너리 빌드
- [ ] ENV-02: 테스트 환경 설정
- [ ] ENV-03: Node 데몬 실행
- [ ] ENV-04: UI 클라이언트 실행

### API 테스트
- [ ] API-01: Health Check
- [ ] API-02: 파일시스템 조회
- [ ] API-03: 경로 조회
- [ ] API-04: 수동 스캔
- [ ] API-05: 프로세스 조회
- [ ] API-06: 로그 조회
- [ ] API-07: Config 조회
- [ ] API-08: Config 수정

### 기본 네비게이션
- [ ] TC-001: 뷰 전환 (0-5)
- [ ] TC-002: Help 뷰
- [ ] TC-003: 자동 새로고침
- [ ] TC-004: 수동 새로고침

### 데이터 표시
- [ ] TC-005: Overview FS
- [ ] TC-006: Overview Path
- [ ] TC-007: Overview 부분 실패
- [ ] TC-008: FS 뷰
- [ ] TC-009: Paths 스캔
- [ ] TC-010: Paths 빈 상태
- [ ] TC-011: Process 색상
- [ ] TC-012: Logs 스크롤
- [ ] TC-013: Logs 빈 상태

### Settings 뷰
- [ ] TC-014: 섹션 네비게이션
- [ ] TC-015: Process 패턴 추가
- [ ] TC-016: Process 패턴 삭제
- [ ] TC-017: Log 추가
- [ ] TC-018: Path 추가
- [ ] TC-019: 설정 저장
- [ ] TC-020: 모달 취소
- [ ] TC-021: 모달 키 보호
- [ ] TC-022: 자동 새로고침 보호
- [ ] TC-023: Dirty 플래그 보호

### 오류 처리
- [ ] TC-024: API 연결 오류
- [ ] TC-025: Settings 저장 오류
- [ ] TC-026: Paths 스캔 오류

### 엣지 케이스
- [ ] TC-027: 빈 응답
- [ ] TC-028: 대용량 데이터
- [ ] TC-029: 긴 값
- [ ] TC-030: 빠른 뷰 전환
- [ ] TC-031: 동시 새로고침
- [ ] TC-032: 다중 모달

### 회귀 테스트
- [ ] TC-033: Settings 모달 포커스
- [ ] TC-034: Settings 자동 새로고침 데이터 손실
- [ ] TC-035: Settings 저장 키 전파

### 성능
- [ ] TC-036: 자동 새로고침 부하
- [ ] TC-037: 대용량 로그

---

**문서 버전:** v2.0
**최종 수정일:** 2026-02-14
**작성자:** Claude Code (Sonnet 4.5)
