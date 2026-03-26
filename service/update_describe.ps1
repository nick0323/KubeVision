@echo off
cd /d %~dp0
powershell -Command "(Get-Content describe.go) -replace 'return &DescribeResult\\{\\s*Name:\\s*\\w+\\.Name,\\s*Namespace:\\s*[^,]*,\\s*Kind:\\s*\"[^\"]*\",\\s*APIVersion:\\s*\"[^\"]*\",\\s*CreatedAt:\\s*\\w+\\.CreationTimestamp\\.Format\\(time\\.RFC3339\\),\\s*Labels:\\s*\\w+\\.Labels,\\s*Annotations:\\s*\\w+\\.Annotations,\\s*Spec:\\s*spec,\\s*Status:\\s*status,\\s*Events:\\s*events,\\s*\\}, nil', 'return &DescribeResult{ Metadata: $1.ObjectMeta, Spec: spec, Status: status, Events: events, }, nil' | Set-Content describe.go"
echo Done!
