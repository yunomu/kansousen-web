#!/bin/sh

PROFILE=default
REGION=us-east-1

if [ $# -lt 1 ]; then
	echo $0 stack_name
	exit 1
fi
STACK=$1

aws --profile $PROFILE --region $REGION cloudformation describe-stacks --stack-name $STACK | jq '.Stacks[].Outputs | map({(.OutputKey):.OutputValue}) | add'
