# DataNexus 微信小游戏登录服务

本项目已新增微信云托管登录接口：

- 请求：`POST /wechat/login`
- 调用方式：微信小游戏 `wx.cloud.callContainer`
- 用户身份：读取微信云托管自动注入的 `X-WX-OPENID`
- 数据库：使用模板自带 MySQL，自动创建 `DatanexusUsers` 表
- 返回：`openId`、`isNewUser`、`serverTime`

## 小游戏调用示例

```javascript
wx.cloud.init();

wx.cloud.callContainer({
    config: {
        env: 'prod-d8gdro2nhb8e0b368'
    },
    path: '/wechat/login',
    method: 'POST',
    header: {
        'X-WX-SERVICE': 'golang-yoy8',
        'content-type': 'application/json'
    },
    data: {},
    success(res) {
        const result = res && res.data;

        if (!result || !result.success || !result.openId) {
            console.error('[WechatLogin] 登录失败', result);
            return;
        }

        GameGlobal.DataNexus.setOpenId(
            result.openId,
            result.isNewUser ? 1 : 0
        );
    },
    fail(error) {
        console.error('[WechatLogin] 云托管调用失败', error);
    }
});
```

## 预期返回

首次登录：

```json
{
  "success": true,
  "openId": "o...",
  "isNewUser": true,
  "serverTime": 1785230000
}
```

之后登录：

```json
{
  "success": true,
  "openId": "o...",
  "isNewUser": false,
  "serverTime": 1785231000
}
```

## 发布

修改文件后提交到当前仓库的 `master` 分支：

```bash
git add .
git commit -m "add DataNexus WeChat login"
git push origin master
```

云托管流水线完成构建后，在服务的运行日志中确认：

```text
[WechatLogin] success, user=xxxxxxxx, isNewUser=true
```

日志只记录 OpenID 的哈希片段，不会打印完整 OpenID。
