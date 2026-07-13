# 香港电商主体支付接入下一步

适用场景：
- 你已经有香港电商主体
- 当前项目后端已接入 `PayPal Sandbox`
- `Stripe` / `连连国际` 还处于接口预留阶段

## 你现在最该先做什么

按优先级建议直接这样推进：

1. 先申请 `PayPal Business`
2. 同步申请 `Stripe Hong Kong`
3. 把项目线上环境变量补齐，先跑通 PayPal Sandbox
4. 再去申请 `连连国际`
5. 等其中任一正式通道通过，再切生产

原因：
- 代码现在已经能直接联调 PayPal
- Stripe 适合后续做网站信用卡主通道
- 连连国际适合做备用收款和跨境结算

## 第一步：整理开户资料包

建议你先准备一个文件夹，名字可以叫 `payment-kyc-pack`，里面放：

- 香港主体登记文件
- 商业登记证 / BR
- 主体名称英文版
- 注册地址英文版
- 负责人身份证 / 护照
- 对应银行账户资料
- 网站域名
- 产品介绍
- 定价说明
- 隐私政策
- 服务条款
- 退款政策
- 联系我们页面

如果你现在网站资料还不完整，至少要先保证前端能公开访问以下页面：

- 首页
- 价格页
- 联系方式
- Privacy Policy
- Terms of Service
- Refund Policy

## 第二步：先跑通 PayPal

当前项目代码已经支持：

- 创建 PayPal Checkout 订单
- 跳转 PayPal
- 回跳前端
- 自动确认支付

你需要做的只有拿到 Sandbox 凭据并填到环境变量里。

### PayPal 要申请什么

你当前阶段不用先纠结生产开户，先拿开发测试资料：

1. 注册 / 登录 PayPal Business
2. 进入 PayPal Developer
3. 创建 Sandbox app
4. 拿到：
   - `Client ID`
   - `Secret`

### 项目里要填的变量

文件位置：
- `backend/.env.example`

线上 Render 也需要配置同名变量：

- `FRONTEND_BASE_URL`
- `PAYPAL_CLIENT_ID`
- `PAYPAL_SECRET`
- `PAYPAL_BASE_URL=https://api-m.sandbox.paypal.com`

### 变量填写示例

```env
FRONTEND_BASE_URL=https://你的-vercel-域名
PAYPAL_CLIENT_ID=你的-paypal-sandbox-client-id
PAYPAL_SECRET=你的-paypal-sandbox-secret
PAYPAL_BASE_URL=https://api-m.sandbox.paypal.com
```

### 代码里已经接好的位置

- `backend/internal/config/config.go`
- `backend/internal/services/paypal_gateway.go`
- `backend/internal/services/payment_service.go`

## 第三步：申请 Stripe Hong Kong

这一条建议你现在就并行申请。

你当前是香港电商主体，先按真实主体类型提交。如果 Stripe 后台要求你选择主体类型：

- 若你是公司，选 `Company`
- 若你是香港个体 / 独资 / sole proprietorship，按后台实际可选项选择 `Sole Proprietorship`

Stripe 官方对香港主体会核验：

- 主体信息
- 负责人身份
- 银行账户
- 网站和业务真实性

要点：
- 所有资料名称必须一致
- 网站上必须看得出你卖什么
- 联系方式和退款条款不要缺

### Stripe 这一步我建议你准备

- 主体英文名
- BR / 商业登记信息
- 银行账户信息
- 负责人证件
- 官网与业务页面

说明：
- 当前项目还没有接 Stripe API
- 现在先做开户，不急着开发
- 一旦你拿到 Stripe 账号，我下一步就可以帮你把 Stripe 网关正式接进后端

## 第四步：申请连连国际

连连国际现在适合当备用和结算补充通道。

但你这里有一个小重点：

- 连连公开文档里对 `香港企业` 和 `香港个人` 是分开说明的
- 你是“香港电商个体主体”，提交时要以后台实际支持的主体类型为准
- 如果注册页不确定，优先联系连连客服确认你应走 `香港企业用户` 还是 `香港个人/个体用户` 路径

建议你准备：

- 商业登记文件
- 主体证件
- 负责人证件
- 银行账户
- 业务协议 / 店铺链接 / 网站链接
- 如有订单或平台流水，也一起备好

说明：
- 当前代码里 `连连国际` 还只是预留网关
- 现在先把商户申请下来最重要

## 第五步：技术上你只需要先完成这一件事

把 PayPal Sandbox 的四个变量准备好给我：

- `FRONTEND_BASE_URL`
- `PAYPAL_CLIENT_ID`
- `PAYPAL_SECRET`
- `PAYPAL_BASE_URL`（默认可直接用 sandbox）

你把这几项给我后，我就可以继续帮你做：

1. 检查本地和线上配置
2. 提交代码并推送
3. 帮你验证 PayPal Sandbox 支付闭环
4. 再进入 Stripe 正式接入

## 最省事的执行顺序

今天开始，直接按这个节奏：

1. 先申请 PayPal Sandbox / Business
2. 并行申请 Stripe Hong Kong
3. 连连国际先提交开户
4. 把 PayPal Sandbox 凭据发给我
5. 我继续完成项目里的线上联调

## 你发我什么，我就继续做

你下一条只要把下面这些发我就行：

```text
FRONTEND_BASE_URL=
PAYPAL_CLIENT_ID=
PAYPAL_SECRET=
```

如果你还没拿到，也可以先告诉我：

- 你准备用哪个正式域名
- 你要先申请 PayPal 还是 Stripe

我下一步就按你当前主体情况继续往下推。
