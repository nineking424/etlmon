# etlmon 수동 테스트 가이드

이 문서는 etlmon MVP를 수동으로 테스트하는 방법을 설명합니다.

## 목차

- [사전 요구사항](#사전-요구사항)
- [1단계: 바이너리 빌드](#1단계-바이너리-빌드)
- [2단계: 테스트 환경 설정](#2단계-테스트-환경-설정)
- [3단계: Node 데몬 실행](#3단계-node-데몬-실행)
- [4단계: API 엔드포인트 테스트](#4단계-api-엔드포인트-테스트)
- [5단계: 테스트 파일 생성](#5단계-테스트-파일-생성)
- [6단계: TUI 클라이언트 실행](#6단계-tui-클라이언트-실행)
- [7단계: TUI 네비게이션](#7단계-tui-네비게이션)
- [8단계: 정리](#8단계-정리)
- [문제 해결](#문제-해결)

---

## 사전 요구사항

- Go 1.24.0 이상
- curl (API 테스트용)
- jq (JSON 출력 포맷팅용, 선택사항)

## 1단계: 바이너리 빌드

프로젝트 루트 디렉토리에서 실행:

```bash
# 프로젝트 디렉토리로 이동
cd /path/to/etlmon

# Node 데몬 빌드
go build -o etlmon-node ./cmd/node

# TUI 클라이언트 빌드
go build -o etlmon-ui ./cmd/ui
```

**예상 결과:**
- `etlmon-node` 바이너리 생성 (~12MB)
- `etlmon-ui` 바이너리 생성 (~10MB)

## 2단계: 테스트 환경 설정

### 2.1 테스트 디렉토리 생성

```bash
mkdir -p /tmp/etlmon/data
```

### 2.2 Node 설정 파일 생성

`/tmp/etlmon/node-test.yaml` 파일 생성:

```yaml
node:
  listen: 127.0.0.1:8080
  node_name: test-node
  db_path: /tmp/etlmon/etlmon.db

refresh:
  disk: 5s
  default_path_scan: 30s

paths:
  - path: /tmp/etlmon/data
    scan_interval: 10s
    max_depth: 3
```

### 2.3 UI 설정 파일 생성

`/tmp/etlmon/ui-test.yaml` 파일 생성:

```yaml
nodes:
  - name: test-node
    address: http://127.0.0.1:8080

ui:
  refresh_interval: 2s
  default_node: test-node
```

## 3단계: Node 데몬 실행

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
API server listening on 127.0.0.1:8080
```

## 4단계: API 엔드포인트 테스트

**터미널 2**에서 curl 명령어로 API 테스트:

### 4.1 헬스 체크

```bash
curl http://127.0.0.1:8080/health
```

**예상 응답:**
```json
{"status":"ok","node_name":"test-node"}
```

### 4.2 파일시스템 사용량 조회

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
      "collected_at": "2026-02-03T22:30:00Z"
    }
  ]
}
```

### 4.3 모니터링 경로 조회

```bash
curl http://127.0.0.1:8080/api/v1/paths | jq
```

**예상 응답:**
```json
{
  "paths": [
    {
      "path": "/tmp/etlmon/data",
      "file_count": 0,
      "dir_count": 0,
      "total_size": 0,
      "last_scan": "2026-02-03T22:30:00Z"
    }
  ]
}
```

### 4.4 수동 스캔 트리거

```bash
curl -X POST http://127.0.0.1:8080/api/v1/paths/scan | jq
```

**예상 응답:**
```json
{"status":"scan_triggered"}
```

## 5단계: 테스트 파일 생성

모니터링할 테스트 파일들을 생성합니다:

```bash
# 테스트 파일 생성
mkdir -p /tmp/etlmon/data/subdir
echo "테스트 파일 1" > /tmp/etlmon/data/file1.txt
echo "테스트 파일 2" > /tmp/etlmon/data/file2.txt
echo "테스트 파일 3" > /tmp/etlmon/data/subdir/file3.txt

# 큰 파일 생성 (선택사항)
dd if=/dev/zero of=/tmp/etlmon/data/large_file.bin bs=1M count=10 2>/dev/null
```

### 5.1 스캔 후 결과 확인

```bash
# 수동 스캔 트리거
curl -X POST http://127.0.0.1:8080/api/v1/paths/scan

# 잠시 대기 후 결과 확인
sleep 2
curl http://127.0.0.1:8080/api/v1/paths | jq
```

**예상 응답:**
```json
{
  "paths": [
    {
      "path": "/tmp/etlmon/data",
      "file_count": 4,
      "dir_count": 1,
      "total_size": 10485810,
      "last_scan": "2026-02-03T22:31:00Z"
    }
  ]
}
```

## 6단계: TUI 클라이언트 실행

### 6.1 설정 파일 사용

**터미널 3**에서 실행:

```bash
./etlmon-ui -c /tmp/etlmon/ui-test.yaml
```

### 6.2 직접 연결 (설정 파일 없이)

```bash
./etlmon-ui --node http://127.0.0.1:8080
```

## 7단계: TUI 네비게이션

TUI가 실행되면 다음 키를 사용하여 네비게이션합니다:

| 키 | 동작 |
|----|------|
| `1` | FS 뷰 (디스크 사용량 표시) |
| `2` | Paths 뷰 (모니터링 경로 표시) |
| `T` | 테이블 포맷 토글 |
| `r` | 강제 새로고침 |
| `s` | 수동 스캔 트리거 |
| `q` | 종료 |
| `Ctrl+C` | 강제 종료 |

### 7.1 FS 뷰 화면 예시

```
┌─────────────────────────────────────────────────────────────┐
│ etlmon - test-node                              [FS View]   │
├─────────────────────────────────────────────────────────────┤
│ Mount Point       Device          Used/Total        Usage   │
│ /                 /dev/disk1s1    234G/494G         47.4%   │
│ /System/Volumes   /dev/disk1s2    12G/494G          2.4%    │
├─────────────────────────────────────────────────────────────┤
│ [1]FS [2]Paths | [r]Refresh [s]Scan [T]Table [q]Quit        │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 Paths 뷰 화면 예시

```
┌─────────────────────────────────────────────────────────────┐
│ etlmon - test-node                           [Paths View]   │
├─────────────────────────────────────────────────────────────┤
│ Path                    Files    Dirs    Size      Scanned  │
│ /tmp/etlmon/data        4        1       10.0MB    22:31:00 │
├─────────────────────────────────────────────────────────────┤
│ [1]FS [2]Paths | [r]Refresh [s]Scan [T]Table [q]Quit        │
└─────────────────────────────────────────────────────────────┘
```

## 8단계: 정리

테스트 완료 후 정리:

```bash
# Node 데몬 종료 (터미널 1에서 Ctrl+C)
# TUI 종료 (터미널 3에서 'q' 또는 Ctrl+C)

# 테스트 파일 삭제
rm -rf /tmp/etlmon
```

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

### 문제: TUI 연결 실패 - "connection refused"

**원인:** Node 데몬이 실행되지 않음

**해결:**
1. Node 데몬이 실행 중인지 확인
2. 주소와 포트가 올바른지 확인
3. 방화벽 설정 확인

### 문제: 경로 스캔 결과가 0

**원인:** 설정된 경로가 존재하지 않거나 권한 없음

**해결:**
```bash
# 경로 존재 확인
ls -la /tmp/etlmon/data

# 권한 확인
stat /tmp/etlmon/data
```

### 문제: 데이터베이스 오류

**원인:** DB 파일 손상 또는 권한 문제

**해결:**
```bash
# DB 파일 삭제 후 재시작
rm /tmp/etlmon/etlmon.db*
./etlmon-node -c /tmp/etlmon/node-test.yaml
```

---

## 빠른 테스트 스크립트

아래 스크립트를 사용하면 전체 테스트 환경을 한 번에 설정할 수 있습니다:

```bash
#!/bin/bash
# quick-test-setup.sh

set -e

# 테스트 디렉토리 생성
mkdir -p /tmp/etlmon/data/subdir

# Node 설정 파일 생성
cat > /tmp/etlmon/node-test.yaml << 'EOF'
node:
  listen: 127.0.0.1:8080
  node_name: test-node
  db_path: /tmp/etlmon/etlmon.db

refresh:
  disk: 5s
  default_path_scan: 30s

paths:
  - path: /tmp/etlmon/data
    scan_interval: 10s
    max_depth: 3
EOF

# UI 설정 파일 생성
cat > /tmp/etlmon/ui-test.yaml << 'EOF'
nodes:
  - name: test-node
    address: http://127.0.0.1:8080

ui:
  refresh_interval: 2s
  default_node: test-node
EOF

# 테스트 파일 생성
echo "테스트 파일 1" > /tmp/etlmon/data/file1.txt
echo "테스트 파일 2" > /tmp/etlmon/data/file2.txt
echo "테스트 파일 3" > /tmp/etlmon/data/subdir/file3.txt

echo "테스트 환경 설정 완료!"
echo ""
echo "다음 명령어로 실행하세요:"
echo "  터미널 1: ./etlmon-node -c /tmp/etlmon/node-test.yaml"
echo "  터미널 2: ./etlmon-ui -c /tmp/etlmon/ui-test.yaml"
```

---

**문서 버전:** 1.0
**최종 수정일:** 2026-02-03
**작성자:** Claude Code (ultrapilot)
