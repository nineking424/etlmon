# etlmon

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-90.1%25-brightgreen.svg)](https://github.com/nineking424/etlmon)
[![CGO Free](https://img.shields.io/badge/CGO-Free-success.svg)](https://github.com/nineking424/etlmon)

Go 기반의 TUI 리소스 모니터링 및 집계 도구입니다. 시스템 메트릭(CPU, 메모리, 디스크)을 수집하고, 설정 가능한 시간 윈도우로 집계하여 SQLite에 저장하며, 인터랙티브 터미널 UI로 데이터를 표시합니다.

## 빠른 시작

```bash
# 클론 및 빌드
git clone https://github.com/nineking424/etlmon.git
cd etlmon
make build-static

# 기본 설정으로 실행
./etlmon --config configs/config.yaml
```

## 주요 기능

| 기능 | 설명 |
|------|------|
| **실시간 모니터링** | 설정 가능한 간격으로 CPU, 메모리, 디스크 사용량 실시간 표시 |
| **시간 윈도우 집계** | 1분, 5분, 1시간 윈도우에 대해 AVG, MAX, MIN, LAST 계산 |
| **영구 저장소** | WAL 모드의 SQLite 데이터베이스로 히스토리 데이터 저장 |
| **인터랙티브 TUI** | 키보드 네비게이션이 가능한 tview 기반 터미널 UI |
| **단일 바이너리** | 외부 런타임 의존성 불필요 |
| **CGO-Free** | 쉬운 크로스 컴파일을 위한 순수 Go SQLite 드라이버 (modernc.org/sqlite) |
| **크로스 플랫폼** | Linux, macOS, Windows용 빌드 지원 |

## 스크린샷

```
┌─ 실시간 메트릭 (Tab으로 전환, Q로 종료) ───────────────────────────┐
│ ━━━ CPU ━━━                                                        │
│   usage_percent      [green]23.5%[white]                           │
│                                                                    │
│ ━━━ MEMORY ━━━                                                     │
│   available_bytes    12.4 GB                                       │
│   total_bytes        16.0 GB                                       │
│   usage_percent      [yellow]77.2%[white]                          │
│   used_bytes         12.4 GB                                       │
│                                                                    │
│ ━━━ DISK ━━━                                                       │
│   usage_percent      [green]45.1%[white] (/)                       │
│                                                                    │
│ 마지막 업데이트: 15:04:05                                           │
└────────────────────────────────────────────────────────────────────┘
│ [green]실행 중[white] | 마지막: 15:04:05 | etlmon                   │
```

## 설치

### 소스에서 빌드 (권장)

```bash
# 저장소 클론
git clone https://github.com/nineking424/etlmon.git
cd etlmon

# 정적 바이너리 빌드 (CGO-free, 권장)
make build-static

# 또는 일반 빌드
make build

# 또는 Linux용 크로스 컴파일
make build-linux
```

### Go Install 사용

```bash
go install github.com/nineking424/etlmon/cmd/etlmon@latest
```

### 사전 빌드된 바이너리

[Releases](https://github.com/nineking424/etlmon/releases)에서 다운로드하세요.

## 사용법

### 명령줄 옵션

```bash
# 설정 파일로 실행 (필수)
./etlmon --config configs/config.yaml

# 데이터베이스 경로 오버라이드
./etlmon --config configs/config.yaml --db /tmp/metrics.db

# 버전 표시
./etlmon --version

# 도움말 표시
./etlmon --help
```

### 키보드 단축키

| 키 | 동작 |
|----|------|
| `Tab` | 실시간/히스토리 뷰 전환 |
| `R` | 실시간 뷰로 전환 |
| `H` | 히스토리 뷰로 전환 |
| `1` | 1분 윈도우로 히스토리 필터 |
| `5` | 5분 윈도우로 히스토리 필터 |
| `0` | 1시간 윈도우로 히스토리 필터 |
| `Q` / `Esc` / `Ctrl+C` | 애플리케이션 종료 |

## 설정

YAML 설정 파일을 생성하세요 (`configs/config.yaml` 예제 참조):

```yaml
# 수집 간격 (메트릭 수집 주기)
interval: 10s

# 모니터링할 리소스
resources:
  - cpu      # CPU 사용률
  - memory   # 메모리 사용량, 가용량, 전체
  - disk     # 마운트 포인트별 디스크 사용량

# 집계 시간 윈도우
windows:
  - 1m       # 1분 집계
  - 5m       # 5분 집계
  - 1h       # 1시간 집계

# 적용할 집계 함수
aggregations:
  - avg      # 윈도우 내 평균값
  - max      # 윈도우 내 최대값
  - min      # 윈도우 내 최소값
  - last     # 윈도우 내 마지막 값

# 데이터베이스 설정
database:
  path: ./etlmon.db  # SQLite 데이터베이스 파일 경로
```

### 설정 옵션 참조

| 옵션 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `interval` | duration | 메트릭 수집 간격 | `10s` |
| `resources` | list | 모니터링할 리소스: `cpu`, `memory`, `disk` | 전체 |
| `windows` | list | 집계 윈도우 (Go duration 형식) | `1m, 5m, 1h` |
| `aggregations` | list | 함수: `avg`, `max`, `min`, `last` | 전체 |
| `database.path` | string | SQLite 데이터베이스 파일 경로 | `./etlmon.db` |

## 아키텍처

```
etlmon/
├── cmd/etlmon/           # 애플리케이션 진입점
│   └── main.go           # CLI 플래그, 컴포넌트 연결, 메인 루프
├── internal/
│   ├── config/           # YAML 설정 파싱 및 검증
│   ├── collector/        # 시스템 메트릭 수집기 (CPU, 메모리, 디스크)
│   ├── aggregator/       # 시간 윈도우 집계 엔진
│   ├── storage/          # WAL 모드의 SQLite DAO
│   └── tui/              # tview 기반 터미널 UI
├── configs/              # 예제 설정 파일
└── testdata/             # 테스트 픽스처
```

### 데이터 흐름

```
┌───────────┐     ┌────────────┐     ┌─────────┐     ┌─────────┐
│ Collector │────▶│ Aggregator │────▶│ Storage │────▶│   TUI   │
│ (CPU/Mem/ │     │ (Windows)  │     │ (SQLite)│     │ (tview) │
│   Disk)   │     └────────────┘     └─────────┘     └─────────┘
└───────────┘           │                                 │
      │                 │                                 │
      └─────────────────┴─────────────────────────────────┘
                    실시간 업데이트
```

1. **Collector** - 설정된 간격으로 원시 메트릭 수집 (메모리에만 저장)
2. **Aggregator** - 시간 윈도우 버퍼 관리, 윈도우 완료 시 집계 계산
3. **Storage** - 완료된 집계만 SQLite에 저장 (원시 메트릭은 저장하지 않음)
4. **TUI** - 실시간 메트릭과 히스토리 집계 데이터 표시

## 개발

### 사전 요구사항

- Go 1.22 이상
- Make (선택사항)

### 빌드 명령어

```bash
make build          # 바이너리 빌드
make build-static   # CGO-free 정적 바이너리 빌드
make build-linux    # Linux amd64용 크로스 컴파일
make build-all      # 모든 플랫폼용 빌드
```

### 테스트 명령어

```bash
make test           # 모든 테스트 실행
make test-race      # 레이스 디텍터로 테스트 실행
make test-cover     # 커버리지 요약과 함께 테스트 실행
make coverage       # HTML 커버리지 리포트 생성
```

### 기타 명령어

```bash
make fmt            # go fmt으로 코드 포맷
make lint           # golangci-lint 실행 (설치된 경우)
make tidy           # go mod tidy 실행
make clean          # 빌드 아티팩트 제거
make run            # 빌드 후 기본 설정으로 실행
make help           # 사용 가능한 모든 명령어 표시
```

### 테스트 커버리지

| 패키지 | 커버리지 | 목표 |
|--------|----------|------|
| `aggregator` | 97.9% | ≥ 90% |
| `collector` | 95.9% | ≥ 70% |
| `config` | 93.6% | ≥ 80% |
| `storage` | 83.0% | ≥ 80% |
| `tui` | 73.0% | ≥ 50% |
| **전체** | **90.1%** | ≥ 75% |

## 데이터베이스 스키마

etlmon은 집계된 메트릭만 SQLite에 저장합니다 (원시 데이터는 저장하지 않음):

```sql
CREATE TABLE aggregated_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,       -- Unix 타임스탬프 (윈도우 종료 시간)
    resource_type TEXT NOT NULL,      -- cpu, memory, disk
    metric_name TEXT NOT NULL,        -- usage_percent, used_bytes 등
    aggregated_value REAL NOT NULL,   -- 계산된 값
    window_size TEXT NOT NULL,        -- 1m, 5m, 1h
    aggregation_type TEXT NOT NULL    -- avg, max, min, last
);

-- 효율적인 쿼리를 위한 인덱스
CREATE INDEX idx_metrics_timestamp ON aggregated_metrics(timestamp);
CREATE INDEX idx_metrics_resource ON aggregated_metrics(resource_type);
CREATE INDEX idx_metrics_window ON aggregated_metrics(window_size);
CREATE INDEX idx_metrics_composite ON aggregated_metrics(resource_type, window_size, timestamp);
```

## 문제 해결

### "Config file required" 오류
```bash
# 설정 파일 경로 지정
./etlmon --config /path/to/config.yaml
```

### 데이터베이스 "Permission denied"
```bash
# 데이터베이스 디렉토리에 쓰기 권한 확인
chmod 755 /path/to/db/directory
```

### 높은 CPU 사용량
- 설정에서 수집 간격 증가 (예: `interval: 30s`)
- 모니터링 리소스 수 감소

### TUI가 제대로 표시되지 않음
- 터미널이 유니코드와 256색상을 지원하는지 확인
- 터미널 창 크기 조절 시도
- 최신 터미널 에뮬레이터 사용 (iTerm2, Alacritty 등)

## 메트릭 참조

### CPU 메트릭
| 메트릭 | 설명 |
|--------|------|
| `usage_percent` | 모든 코어의 평균 CPU 사용률 (0-100) |

### 메모리 메트릭
| 메트릭 | 설명 |
|--------|------|
| `usage_percent` | 메모리 사용률 (0-100) |
| `used_bytes` | 사용 중인 메모리 바이트 |
| `available_bytes` | 사용 가능한 메모리 바이트 |
| `total_bytes` | 전체 시스템 메모리 |

### 디스크 메트릭
| 메트릭 | 설명 |
|--------|------|
| `usage_percent` | 마운트 포인트별 디스크 사용률 (0-100) |
| `used_bytes` | 디스크에서 사용된 바이트 |
| `total_bytes` | 전체 디스크 용량 |

참고: 가상 파일시스템(tmpfs, proc, sysfs 등)은 자동으로 필터링됩니다.

## 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 기여하기

1. 저장소 포크
2. 기능 브랜치 생성 (`git checkout -b feature/amazing-feature`)
3. 테스트 먼저 작성 (TDD 방식)
4. 변경사항 구현
5. 모든 테스트 통과 확인 (`make test`)
6. 변경사항 커밋 (`git commit -m 'Add amazing feature'`)
7. 브랜치에 푸시 (`git push origin feature/amazing-feature`)
8. Pull Request 생성

## 감사의 글

- [tview](https://github.com/rivo/tview) - 터미널 UI 라이브러리
- [modernc.org/sqlite](https://modernc.org/sqlite) - CGO-free SQLite 드라이버
- [gopsutil](https://github.com/shirou/gopsutil) - 시스템 메트릭 수집
