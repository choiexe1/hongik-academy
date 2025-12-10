#!/bin/bash

# 환경 설정 파일 로드
if [ ! -f "./deploy.env" ]; then
    echo "deploy.env 파일이 없습니다!"
    echo "deploy.env.example을 참고하여 deploy.env를 생성하세요."
    exit 1
fi
source ./deploy.env

REMOTE_DIR="/home/$SERVER_USER/academy"

echo "=== 홍익미술학원 배포 스크립트 ==="

# 1. 바이너리 빌드 (AMD64)
echo "[1/4] 바이너리 빌드 중..."
GOOS=linux GOARCH=amd64 go build -o academy ./cmd/server
if [ $? -ne 0 ]; then
    echo "빌드 실패!"
    exit 1
fi
echo "빌드 완료: academy"

# 2. 배포 파일 압축
echo "[2/4] 파일 압축 중..."
tar czf deploy.tar.gz academy templates migrations docker-compose.yaml
echo "압축 완료: deploy.tar.gz"

# 3. 서버로 전송
echo "[3/4] 서버로 전송 중..."
scp -i $SSH_KEY -o StrictHostKeyChecking=no deploy.tar.gz $SERVER_USER@$SERVER_IP:/home/$SERVER_USER/
if [ $? -ne 0 ]; then
    echo "전송 실패!"
    exit 1
fi
echo "전송 완료"

# 4. 서버에서 압축 해제, 서비스 설정 및 재시작
echo "[4/4] 서버 배포 및 재시작 중..."
ssh -i $SSH_KEY -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP \
    "DB_USER='$DB_USER' DB_PASSWORD='$DB_PASSWORD' DB_NAME='$DB_NAME' SESSION_KEY='$SESSION_KEY'" \
    bash << 'EOF'
    set -e

    # 파일 압축 해제
    mkdir -p ~/academy
    cd ~/academy
    tar xzf ~/deploy.tar.gz
    rm ~/deploy.tar.gz
    chmod +x academy

    # systemd 서비스 파일 생성/업데이트
    echo "systemd 서비스 설정 중..."
    sudo tee /etc/systemd/system/academy.service > /dev/null << SERVICE
[Unit]
Description=Hongik Academy Management System
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/academy
Environment=TZ=Asia/Seoul
Environment=DB_USER=$DB_USER
Environment=DB_PASSWORD=$DB_PASSWORD
Environment=DB_NAME=$DB_NAME
Environment=SESSION_KEY=$SESSION_KEY
ExecStart=/home/ubuntu/academy/academy
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SERVICE
    sudo systemctl daemon-reload
    sudo systemctl enable academy

    # DB 컨테이너 확인 및 시작
    cd ~/academy
    if ! sudo docker ps | grep -q academy-db; then
        echo "DB 컨테이너 시작 중..."
        sudo docker-compose up -d
        sleep 3
    fi

    # 기존 프로세스 종료 (8080 포트 사용 중인 모든 프로세스)
    echo "기존 프로세스 정리 중..."
    sudo systemctl stop academy 2>/dev/null || true
    sudo fuser -k 8080/tcp 2>/dev/null || true
    sleep 2

    # 서비스 재시작
    echo "앱 재시작 중..."
    sudo systemctl restart academy
    sleep 2

    # 상태 확인
    if sudo systemctl is-active --quiet academy; then
        echo "✓ 앱이 정상적으로 실행 중입니다!"
        sudo systemctl status academy --no-pager | head -10
    else
        echo "✗ 앱 시작 실패. 로그 확인:"
        sudo journalctl -u academy -n 20 --no-pager
        exit 1
    fi
EOF

# 로컬 정리
rm -f deploy.tar.gz academy
echo ""
echo "=== 배포 완료 ==="
echo ""
echo "접속 주소: http://$SERVER_IP:8080"
echo ""
echo "유용한 명령어 (서버에서):"
echo "  sudo systemctl status academy    # 상태 확인"
echo "  sudo systemctl restart academy   # 재시작"
echo "  sudo journalctl -u academy -f    # 로그 보기"
