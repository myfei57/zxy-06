# GroundOps

GroundOps 是机场地勤保障调度平台：航班动态驱动保障任务分配，任务执行占用人员与
车辆资源，超时按 SLA 升级，班组换班交接在途任务，异常后恢复补扫，全部操作留审计。

## 功能

- 航班注册与动态更新、取消级联
- 保障任务创建、分配、开始、完成、回执
- 人员/车辆资源注册、占用与释放
- SLA 保障窗口计时与超时升级
- 班组排班、交接班与在途任务归属迁移
- 恢复补扫与操作审计
- 控制台页面与 HTTP API

## 运行

```bash
go build -mod=vendor -o groundops ./cmd/server
./groundops -addr :8080 -data ./data
```

打开 http://localhost:8080/ 查看总览，/console/flights、/console/tasks、
/console/resources、/console/audit 分别查看航班、任务、资源与审计页面。

## 测试

```bash
go test -mod=vendor ./...
```
