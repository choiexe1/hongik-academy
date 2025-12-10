#!/bin/bash
source ./deploy.env 2>/dev/null || { echo "deploy.env 파일이 없습니다!"; exit 1; }
ssh -i $SSH_KEY $SERVER_USER@$SERVER_IP
