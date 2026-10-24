// UI translations for the ten most spoken languages of the world.
// Missing keys fall back to English.

export const LANGS = [
  { code: "en", name: "English" },
  { code: "zh", name: "中文" },
  { code: "hi", name: "हिन्दी" },
  { code: "es", name: "Español" },
  { code: "fr", name: "Français" },
  { code: "ar", name: "العربية", rtl: true },
  { code: "bn", name: "বাংলা" },
  { code: "pt", name: "Português" },
  { code: "ru", name: "Русский" },
  { code: "ur", name: "اردو", rtl: true },
];

const D = {
  en: {
    nav: { dashboard: "Dashboard", markets: "Markets", alerts: "Alerts", settings: "Settings" },
    conn: { live: "Live", offline: "Offline", connecting: "Connecting…" },
    section: { crypto: "Cryptocurrencies", fiat: "Fiat currencies" },
    dash: { subtitle: "Real-time prices from {n} exchanges and sources", arbitrage: "Best arbitrage now", services: "Microservices online", alerts: "Active alerts" },
    card: { buy: "Buy", sell: "Sell", sources: "{n} of {m} sources", waiting: "Waiting for data…", down: "Service unavailable", vsNote: "in {vs}" },
    time: { now: "just now", sec: "{n} s ago", min: "{n} min ago", hour: "{n} h ago", day: "{n} d ago" },
    range: { "1h": "1H", "6h": "6H", "24h": "24H", "7d": "7D", "30d": "30D", "90d": "90D", "1y": "1Y", "5y": "5Y" },
    detail: {
      back: "Back", aggregate: "Aggregate", percent: "% change", area: "Area", line: "Line",
      best: "Best rates", bestBuy: "Best place to buy", bestSell: "Best place to sell", arbitrage: "Arbitrage after fees",
      arbHint: "Buy on {buy}, sell on {sell}", noArb: "No profitable arbitrage right now", saving: "Savings",
      savingHint: "Best buy vs. the most expensive source", converter: "Exchange calculator", amount: "Amount",
      buy: "Buy", sell: "Sell", cost: "You pay", receive: "You receive", bestAt: "Best: {src}", worstAt: "Worst: {src}",
      diff: "Difference", sources: "Exchanges & sources", createAlert: "Create alert", noHistory: "No history for this period yet",
      loading: "Loading…", chartHint: "Click sources to compare them on the chart",
    },
    col: { source: "Source", pair: "Pair", bid: "Bid", ask: "Ask", rate: "Rate", spread: "Spread", fee: "Fee", effBuy: "Buy incl. fee", effSell: "Sell incl. fee", change: "24h", updated: "Updated", status: "Status", chart: "Chart" },
    state: { ok: "OK", error: "Error", waiting: "Waiting", excluded: "Excluded", na: "Not quoted" },
    type: { exchange: "Exchange", reference: "Reference rate" },
    markets: { title: "Market matrix", hint: "Live prices on every exchange. Green — cheapest to buy, blue — best to sell (fees included).", compare: "Performance comparison", compareHint: "Change in % over the selected period", fiat: "Fiat rates by source" },
    alerts: {
      title: "Price alerts", new: "New alert", edit: "Edit alert", empty: "No alerts yet. Create one to be notified when a price moves.",
      history: "Notification history", clear: "Clear", noEvents: "No notifications yet", asset: "Currency", vs: "Priced in",
      source: "Source", condition: "Condition", value: "Value", window: "Period", repeat: "Repeat after each notification",
      note: "Note", notePh: "Optional comment", save: "Save", cancel: "Cancel", delete: "Delete", deleteConfirm: "Delete this alert?",
      enable: "Enable", disable: "Pause", active: "Active", paused: "Paused", fired: "Fired {n}×", now: "Now {v}",
      current: "Current price: {v}", fireAt: "Fires at {v}", baseline: "Reference: {v}",
    },
    cond: { rise_pct: "Rises by %", fall_pct: "Falls by %", move_pct: "Moves by ±%", rise_abs: "Rises by amount", fall_abs: "Falls by amount", above: "Goes above", below: "Goes below" },
    win: { since: "Since the alert was created", "5m": "Within 5 minutes", "15m": "Within 15 minutes", "1h": "Within 1 hour", "4h": "Within 4 hours", "24h": "Within 24 hours", "7d": "Within 7 days" },
    msg: { up: "{asset} rose by {chg} {window}", down: "{asset} fell by {chg} {window}", move: "{asset} changed by {chg} {window}", above: "{asset} rose above {level}", below: "{asset} fell below {level}", since: "since the alert was set", in: "in {w}" },
    unit: { min: "min", h: "h", d: "d" },
    settings: {
      title: "Settings", appearance: "Appearance", language: "Language", theme: "Theme", scale: "Interface scale",
      scaleAuto: "Auto (by window size and resolution)", chartStyle: "Chart style", currency: "Display currency",
      notifications: "Notifications", osNotify: "Desktop notifications",
      osNotifyHint: "Native Windows, macOS or Linux notifications — they work even when this window is closed",
      browserNotify: "Browser notifications", sound: "Sound", test: "Send a test notification", testSent: "Test notification sent",
      permDenied: "Notifications are blocked by the browser", sources: "Exchanges and fees",
      sourcesHint: "Unchecked sources are ignored when searching for the best rate. Fees are taker fees used for effective prices.",
      include: "Use", reset: "Reset", services: "Microservices", about: "About",
      aboutText: "Free and open source. Data comes from public exchange APIs and central banks — no API keys or paid services.",
      quit: "Quit Cryptomon", quitConfirm: "Stop all services and quit?", stopped: "Cryptomon has stopped. You can close this window.",
    },
    theme: { system: "System", light: "Light", dark: "Dark", midnight: "Midnight", ocean: "Ocean", sepia: "Sepia" },
    svc: { up: "Running", down: "Unavailable", waiting: "Starting" },
    toast: { saved: "Saved", error: "Error: {e}" },
    asset: { BTC: "Bitcoin", ETH: "Ethereum", USDT: "Tether", TON: "Toncoin", USD: "US dollar", EUR: "Euro", RUB: "Russian ruble", CNY: "Chinese yuan", JPY: "Japanese yen", INR: "Indian rupee", CHF: "Swiss franc", GBP: "British pound" },
  },

  ru: {
    nav: { dashboard: "Обзор", markets: "Рынки", alerts: "Уведомления", settings: "Настройки" },
    conn: { live: "В эфире", offline: "Нет связи", connecting: "Подключение…" },
    section: { crypto: "Криптовалюты", fiat: "Фиатные валюты" },
    dash: { subtitle: "Курсы в реальном времени из {n} бирж и источников", arbitrage: "Лучший арбитраж сейчас", services: "Микросервисы онлайн", alerts: "Активные уведомления" },
    card: { buy: "Купить", sell: "Продать", sources: "{n} из {m} источников", waiting: "Ожидание данных…", down: "Сервис недоступен", vsNote: "в {vs}" },
    time: { now: "только что", sec: "{n} с назад", min: "{n} мин назад", hour: "{n} ч назад", day: "{n} дн назад" },
    range: { "1h": "1Ч", "6h": "6Ч", "24h": "24Ч", "7d": "7Д", "30d": "30Д", "90d": "90Д", "1y": "1Г", "5y": "5Л" },
    detail: {
      back: "Назад", aggregate: "Сводный курс", percent: "Изменение, %", area: "Область", line: "Линия",
      best: "Лучшие курсы", bestBuy: "Где выгоднее купить", bestSell: "Где выгоднее продать", arbitrage: "Арбитраж с учётом комиссий",
      arbHint: "Купить на {buy}, продать на {sell}", noArb: "Сейчас нет выгодного арбитража", saving: "Экономия",
      savingHint: "Лучшая покупка против самого дорогого источника", converter: "Калькулятор обмена", amount: "Количество",
      buy: "Купить", sell: "Продать", cost: "Вы заплатите", receive: "Вы получите", bestAt: "Лучше всего: {src}", worstAt: "Хуже всего: {src}",
      diff: "Разница", sources: "Биржи и источники", createAlert: "Создать уведомление", noHistory: "Нет истории за этот период",
      loading: "Загрузка…", chartHint: "Выберите источники, чтобы сравнить их на графике",
    },
    col: { source: "Источник", pair: "Пара", bid: "Bid", ask: "Ask", rate: "Курс", spread: "Спред", fee: "Комиссия", effBuy: "Покупка с комиссией", effSell: "Продажа с комиссией", change: "24ч", updated: "Обновлено", status: "Статус", chart: "График" },
    state: { ok: "OK", error: "Ошибка", waiting: "Ожидание", excluded: "Исключён", na: "Не котируется" },
    type: { exchange: "Биржа", reference: "Справочный курс" },
    markets: { title: "Матрица рынков", hint: "Текущие цены на всех биржах. Зелёный — где дешевле купить, синий — где выгоднее продать (с учётом комиссий).", compare: "Сравнение динамики", compareHint: "Изменение в % за выбранный период", fiat: "Курсы фиатных валют по источникам" },
    alerts: {
      title: "Ценовые уведомления", new: "Новое уведомление", edit: "Изменить уведомление", empty: "Уведомлений пока нет. Создайте, чтобы узнать о движении цены.",
      history: "История уведомлений", clear: "Очистить", noEvents: "Уведомлений ещё не было", asset: "Валюта", vs: "В валюте",
      source: "Источник", condition: "Условие", value: "Значение", window: "Период", repeat: "Повторять после каждого срабатывания",
      note: "Заметка", notePh: "Необязательный комментарий", save: "Сохранить", cancel: "Отмена", delete: "Удалить", deleteConfirm: "Удалить это уведомление?",
      enable: "Включить", disable: "Приостановить", active: "Активно", paused: "Приостановлено", fired: "Сработало {n}×", now: "Сейчас {v}",
      current: "Текущая цена: {v}", fireAt: "Сработает при {v}", baseline: "Точка отсчёта: {v}",
    },
    cond: { rise_pct: "Вырастет на %", fall_pct: "Упадёт на %", move_pct: "Изменится на ±%", rise_abs: "Вырастет на сумму", fall_abs: "Упадёт на сумму", above: "Поднимется выше", below: "Опустится ниже" },
    win: { since: "С момента создания", "5m": "За 5 минут", "15m": "За 15 минут", "1h": "За 1 час", "4h": "За 4 часа", "24h": "За 24 часа", "7d": "За 7 дней" },
    msg: { up: "{asset} вырос на {chg} {window}", down: "{asset} упал на {chg} {window}", move: "{asset} изменился на {chg} {window}", above: "{asset} поднялся выше {level}", below: "{asset} опустился ниже {level}", since: "с момента создания уведомления", in: "за {w}" },
    unit: { min: "мин", h: "ч", d: "дн" },
    settings: {
      title: "Настройки", appearance: "Оформление", language: "Язык", theme: "Тема", scale: "Масштаб интерфейса",
      scaleAuto: "Авто (по размеру окна и разрешению)", chartStyle: "Вид графика", currency: "Валюта отображения",
      notifications: "Уведомления", osNotify: "Системные уведомления",
      osNotifyHint: "Нативные уведомления Windows, macOS или Linux — работают, даже когда окно закрыто",
      browserNotify: "Уведомления браузера", sound: "Звук", test: "Отправить тестовое уведомление", testSent: "Тестовое уведомление отправлено",
      permDenied: "Браузер заблокировал уведомления", sources: "Биржи и комиссии",
      sourcesHint: "Источники без галочки не участвуют в поиске лучшего курса. Комиссии — тейкерские, используются для расчёта эффективной цены.",
      include: "Учитывать", reset: "Сбросить", services: "Микросервисы", about: "О программе",
      aboutText: "Бесплатно и с открытым кодом. Данные — из публичных API бирж и центробанков, без ключей и платных сервисов.",
      quit: "Выйти из Cryptomon", quitConfirm: "Остановить все сервисы и выйти?", stopped: "Cryptomon остановлен. Окно можно закрыть.",
    },
    theme: { system: "Системная", light: "Светлая", dark: "Тёмная", midnight: "Полночь", ocean: "Океан", sepia: "Сепия" },
    svc: { up: "Работает", down: "Недоступен", waiting: "Запускается" },
    toast: { saved: "Сохранено", error: "Ошибка: {e}" },
    asset: { BTC: "Биткоин", ETH: "Эфириум", USDT: "Tether", TON: "Toncoin", USD: "Доллар США", EUR: "Евро", RUB: "Российский рубль", CNY: "Китайский юань", JPY: "Японская иена", INR: "Индийская рупия", CHF: "Швейцарский франк", GBP: "Британский фунт" },
  },

  zh: {
    nav: { dashboard: "概览", markets: "市场", alerts: "提醒", settings: "设置" },
    conn: { live: "实时", offline: "离线", connecting: "连接中…" },
    section: { crypto: "加密货币", fiat: "法定货币" },
    dash: { subtitle: "来自 {n} 个交易所和数据源的实时价格", arbitrage: "当前最佳套利", services: "在线微服务", alerts: "活跃提醒" },
    card: { buy: "买入", sell: "卖出", sources: "{n}/{m} 个数据源", waiting: "等待数据…", down: "服务不可用", vsNote: "以 {vs} 计" },
    time: { now: "刚刚", sec: "{n} 秒前", min: "{n} 分钟前", hour: "{n} 小时前", day: "{n} 天前" },
    range: { "1h": "1小时", "6h": "6小时", "24h": "24小时", "7d": "7天", "30d": "30天", "90d": "90天", "1y": "1年", "5y": "5年" },
    detail: {
      back: "返回", aggregate: "综合价格", percent: "涨跌 %", area: "面积图", line: "折线图",
      best: "最佳汇率", bestBuy: "最佳买入地", bestSell: "最佳卖出地", arbitrage: "扣除手续费后的套利",
      arbHint: "在 {buy} 买入，在 {sell} 卖出", noArb: "目前没有可盈利的套利机会", saving: "节省",
      savingHint: "最佳买价与最贵数据源相比", converter: "兑换计算器", amount: "数量",
      buy: "买入", sell: "卖出", cost: "您需支付", receive: "您将收到", bestAt: "最佳：{src}", worstAt: "最差：{src}",
      diff: "差额", sources: "交易所和数据源", createAlert: "创建提醒", noHistory: "该时段暂无历史数据",
      loading: "加载中…", chartHint: "点击数据源以在图表上比较",
    },
    col: { source: "数据源", pair: "交易对", bid: "买价", ask: "卖价", rate: "汇率", spread: "价差", fee: "手续费", effBuy: "含费买入价", effSell: "含费卖出价", change: "24小时", updated: "更新时间", status: "状态", chart: "图表" },
    state: { ok: "正常", error: "错误", waiting: "等待中", excluded: "已排除", na: "无报价" },
    type: { exchange: "交易所", reference: "参考汇率" },
    markets: { title: "市场矩阵", hint: "各交易所的实时价格。绿色——最便宜的买入处，蓝色——最佳卖出处（含手续费）。", compare: "走势比较", compareHint: "所选时段内的涨跌幅（%）", fiat: "各数据源的法币汇率" },
    alerts: {
      title: "价格提醒", new: "新建提醒", edit: "编辑提醒", empty: "暂无提醒。创建提醒，在价格变动时获得通知。",
      history: "通知历史", clear: "清除", noEvents: "暂无通知", asset: "货币", vs: "计价货币",
      source: "数据源", condition: "条件", value: "数值", window: "时间段", repeat: "每次通知后重复",
      note: "备注", notePh: "可选备注", save: "保存", cancel: "取消", delete: "删除", deleteConfirm: "删除此提醒？",
      enable: "启用", disable: "暂停", active: "活跃", paused: "已暂停", fired: "已触发 {n} 次", now: "当前 {v}",
      current: "当前价格：{v}", fireAt: "触发价：{v}", baseline: "基准：{v}",
    },
    cond: { rise_pct: "上涨 %", fall_pct: "下跌 %", move_pct: "涨跌 ±%", rise_abs: "上涨金额", fall_abs: "下跌金额", above: "高于", below: "低于" },
    win: { since: "自创建提醒以来", "5m": "5分钟内", "15m": "15分钟内", "1h": "1小时内", "4h": "4小时内", "24h": "24小时内", "7d": "7天内" },
    msg: { up: "{asset} {window}上涨了 {chg}", down: "{asset} {window}下跌了 {chg}", move: "{asset} {window}变动了 {chg}", above: "{asset} 涨破 {level}", below: "{asset} 跌破 {level}", since: "自设置提醒以来", in: "在{w}内" },
    unit: { min: "分钟", h: "小时", d: "天" },
    settings: {
      title: "设置", appearance: "外观", language: "语言", theme: "主题", scale: "界面缩放",
      scaleAuto: "自动（根据窗口大小和分辨率）", chartStyle: "图表样式", currency: "显示货币",
      notifications: "通知", osNotify: "桌面通知",
      osNotifyHint: "Windows、macOS 或 Linux 原生通知——即使关闭此窗口也能收到",
      browserNotify: "浏览器通知", sound: "声音", test: "发送测试通知", testSent: "测试通知已发送",
      permDenied: "浏览器已阻止通知", sources: "交易所和手续费",
      sourcesHint: "未勾选的数据源不参与最佳汇率计算。手续费为吃单手续费，用于计算实际价格。",
      include: "使用", reset: "重置", services: "微服务", about: "关于",
      aboutText: "免费开源。数据来自交易所公开 API 和中央银行——无需 API 密钥或付费服务。",
      quit: "退出 Cryptomon", quitConfirm: "停止所有服务并退出？", stopped: "Cryptomon 已停止。您可以关闭此窗口。",
    },
    theme: { system: "跟随系统", light: "浅色", dark: "深色", midnight: "午夜", ocean: "海洋", sepia: "复古" },
    svc: { up: "运行中", down: "不可用", waiting: "启动中" },
    toast: { saved: "已保存", error: "错误：{e}" },
    asset: { BTC: "比特币", ETH: "以太坊", USDT: "泰达币", TON: "Toncoin", USD: "美元", EUR: "欧元", RUB: "俄罗斯卢布", CNY: "人民币", JPY: "日元", INR: "印度卢比", CHF: "瑞士法郎", GBP: "英镑" },
  },

  hi: {
    nav: { dashboard: "डैशबोर्ड", markets: "बाज़ार", alerts: "अलर्ट", settings: "सेटिंग्स" },
    conn: { live: "लाइव", offline: "ऑफ़लाइन", connecting: "कनेक्ट हो रहा है…" },
    section: { crypto: "क्रिप्टोकरेंसी", fiat: "फ़िएट मुद्राएँ" },
    dash: { subtitle: "{n} एक्सचेंजों और स्रोतों से रीयल-टाइम कीमतें", arbitrage: "अभी सबसे अच्छा आर्बिट्राज", services: "ऑनलाइन माइक्रोसर्विस", alerts: "सक्रिय अलर्ट" },
    card: { buy: "ख़रीदें", sell: "बेचें", sources: "{m} में से {n} स्रोत", waiting: "डेटा की प्रतीक्षा…", down: "सेवा उपलब्ध नहीं", vsNote: "{vs} में" },
    time: { now: "अभी-अभी", sec: "{n} सेकंड पहले", min: "{n} मिनट पहले", hour: "{n} घंटे पहले", day: "{n} दिन पहले" },
    range: { "1h": "1घं", "6h": "6घं", "24h": "24घं", "7d": "7दि", "30d": "30दि", "90d": "90दि", "1y": "1व", "5y": "5व" },
    detail: {
      back: "वापस", aggregate: "समेकित मूल्य", percent: "% बदलाव", area: "क्षेत्र", line: "रेखा",
      best: "सर्वोत्तम दरें", bestBuy: "ख़रीदने के लिए सबसे अच्छी जगह", bestSell: "बेचने के लिए सबसे अच्छी जगह", arbitrage: "शुल्क के बाद आर्बिट्राज",
      arbHint: "{buy} पर ख़रीदें, {sell} पर बेचें", noArb: "अभी कोई लाभदायक आर्बिट्राज नहीं", saving: "बचत",
      savingHint: "सबसे अच्छी ख़रीद बनाम सबसे महँगा स्रोत", converter: "विनिमय कैलकुलेटर", amount: "मात्रा",
      buy: "ख़रीदें", sell: "बेचें", cost: "आप भुगतान करेंगे", receive: "आपको मिलेगा", bestAt: "सबसे अच्छा: {src}", worstAt: "सबसे ख़राब: {src}",
      diff: "अंतर", sources: "एक्सचेंज और स्रोत", createAlert: "अलर्ट बनाएँ", noHistory: "इस अवधि का इतिहास अभी उपलब्ध नहीं",
      loading: "लोड हो रहा है…", chartHint: "चार्ट पर तुलना करने के लिए स्रोत चुनें",
    },
    col: { source: "स्रोत", pair: "जोड़ी", bid: "बोली", ask: "पूछ मूल्य", rate: "दर", spread: "स्प्रेड", fee: "शुल्क", effBuy: "शुल्क सहित ख़रीद", effSell: "शुल्क सहित बिक्री", change: "24घं", updated: "अपडेट", status: "स्थिति", chart: "चार्ट" },
    state: { ok: "ठीक", error: "त्रुटि", waiting: "प्रतीक्षा", excluded: "बाहर रखा गया", na: "उद्धृत नहीं" },
    type: { exchange: "एक्सचेंज", reference: "संदर्भ दर" },
    markets: { title: "बाज़ार मैट्रिक्स", hint: "हर एक्सचेंज पर लाइव कीमतें। हरा — ख़रीदने के लिए सबसे सस्ता, नीला — बेचने के लिए सबसे अच्छा (शुल्क सहित)।", compare: "प्रदर्शन तुलना", compareHint: "चुनी गई अवधि में % बदलाव", fiat: "स्रोत के अनुसार फ़िएट दरें" },
    alerts: {
      title: "मूल्य अलर्ट", new: "नया अलर्ट", edit: "अलर्ट संपादित करें", empty: "अभी कोई अलर्ट नहीं। कीमत बदलने पर सूचना पाने के लिए एक अलर्ट बनाएँ।",
      history: "सूचना इतिहास", clear: "साफ़ करें", noEvents: "अभी कोई सूचना नहीं", asset: "मुद्रा", vs: "मूल्य मुद्रा",
      source: "स्रोत", condition: "शर्त", value: "मान", window: "अवधि", repeat: "हर सूचना के बाद दोहराएँ",
      note: "नोट", notePh: "वैकल्पिक टिप्पणी", save: "सहेजें", cancel: "रद्द करें", delete: "हटाएँ", deleteConfirm: "यह अलर्ट हटाएँ?",
      enable: "चालू करें", disable: "रोकें", active: "सक्रिय", paused: "रुका हुआ", fired: "{n} बार ट्रिगर हुआ", now: "अभी {v}",
      current: "वर्तमान मूल्य: {v}", fireAt: "{v} पर ट्रिगर होगा", baseline: "संदर्भ: {v}",
    },
    cond: { rise_pct: "% बढ़े", fall_pct: "% घटे", move_pct: "±% बदले", rise_abs: "राशि से बढ़े", fall_abs: "राशि से घटे", above: "इससे ऊपर जाए", below: "इससे नीचे जाए" },
    win: { since: "अलर्ट बनाने के बाद से", "5m": "5 मिनट में", "15m": "15 मिनट में", "1h": "1 घंटे में", "4h": "4 घंटे में", "24h": "24 घंटे में", "7d": "7 दिन में" },
    msg: { up: "{asset} {window} {chg} बढ़ा", down: "{asset} {window} {chg} गिरा", move: "{asset} {window} {chg} बदला", above: "{asset} {level} से ऊपर गया", below: "{asset} {level} से नीचे गया", since: "अलर्ट सेट करने के बाद से", in: "{w} में" },
    unit: { min: "मिनट", h: "घंटे", d: "दिन" },
    settings: {
      title: "सेटिंग्स", appearance: "रूप-रंग", language: "भाषा", theme: "थीम", scale: "इंटरफ़ेस स्केल",
      scaleAuto: "स्वचालित (विंडो आकार और रिज़ॉल्यूशन के अनुसार)", chartStyle: "चार्ट शैली", currency: "प्रदर्शन मुद्रा",
      notifications: "सूचनाएँ", osNotify: "डेस्कटॉप सूचनाएँ",
      osNotifyHint: "Windows, macOS या Linux की मूल सूचनाएँ — यह विंडो बंद होने पर भी काम करती हैं",
      browserNotify: "ब्राउज़र सूचनाएँ", sound: "ध्वनि", test: "परीक्षण सूचना भेजें", testSent: "परीक्षण सूचना भेजी गई",
      permDenied: "ब्राउज़र ने सूचनाएँ अवरुद्ध कर दी हैं", sources: "एक्सचेंज और शुल्क",
      sourcesHint: "बिना चिह्न वाले स्रोत सर्वोत्तम दर की खोज में शामिल नहीं होते। शुल्क टेकर शुल्क हैं, जिनसे प्रभावी मूल्य निकाला जाता है।",
      include: "उपयोग", reset: "रीसेट", services: "माइक्रोसर्विस", about: "परिचय",
      aboutText: "मुफ़्त और ओपन सोर्स। डेटा एक्सचेंजों के सार्वजनिक API और केंद्रीय बैंकों से — किसी API कुंजी या सशुल्क सेवा की ज़रूरत नहीं।",
      quit: "Cryptomon बंद करें", quitConfirm: "सभी सेवाएँ रोककर बाहर निकलें?", stopped: "Cryptomon बंद हो गया। आप यह विंडो बंद कर सकते हैं।",
    },
    theme: { system: "सिस्टम", light: "हल्की", dark: "गहरी", midnight: "मध्यरात्रि", ocean: "सागर", sepia: "सेपिया" },
    svc: { up: "चल रहा है", down: "उपलब्ध नहीं", waiting: "शुरू हो रहा है" },
    toast: { saved: "सहेजा गया", error: "त्रुटि: {e}" },
    asset: { BTC: "बिटकॉइन", ETH: "एथेरियम", USDT: "टेदर", TON: "टोनकॉइन", USD: "अमेरिकी डॉलर", EUR: "यूरो", RUB: "रूसी रूबल", CNY: "चीनी युआन", JPY: "जापानी येन", INR: "भारतीय रुपया", CHF: "स्विस फ़्रैंक", GBP: "ब्रिटिश पाउंड" },
  },

  es: {
    nav: { dashboard: "Panel", markets: "Mercados", alerts: "Alertas", settings: "Ajustes" },
    conn: { live: "En vivo", offline: "Sin conexión", connecting: "Conectando…" },
    section: { crypto: "Criptomonedas", fiat: "Monedas fiduciarias" },
    dash: { subtitle: "Precios en tiempo real de {n} exchanges y fuentes", arbitrage: "Mejor arbitraje ahora", services: "Microservicios en línea", alerts: "Alertas activas" },
    card: { buy: "Comprar", sell: "Vender", sources: "{n} de {m} fuentes", waiting: "Esperando datos…", down: "Servicio no disponible", vsNote: "en {vs}" },
    time: { now: "ahora mismo", sec: "hace {n} s", min: "hace {n} min", hour: "hace {n} h", day: "hace {n} d" },
    range: { "1h": "1H", "6h": "6H", "24h": "24H", "7d": "7D", "30d": "30D", "90d": "90D", "1y": "1A", "5y": "5A" },
    detail: {
      back: "Volver", aggregate: "Precio agregado", percent: "Cambio %", area: "Área", line: "Línea",
      best: "Mejores tipos", bestBuy: "Dónde comprar mejor", bestSell: "Dónde vender mejor", arbitrage: "Arbitraje tras comisiones",
      arbHint: "Comprar en {buy}, vender en {sell}", noArb: "Ahora no hay arbitraje rentable", saving: "Ahorro",
      savingHint: "Mejor compra frente a la fuente más cara", converter: "Calculadora de cambio", amount: "Cantidad",
      buy: "Comprar", sell: "Vender", cost: "Pagas", receive: "Recibes", bestAt: "Mejor: {src}", worstAt: "Peor: {src}",
      diff: "Diferencia", sources: "Exchanges y fuentes", createAlert: "Crear alerta", noHistory: "Aún no hay historial para este periodo",
      loading: "Cargando…", chartHint: "Haz clic en las fuentes para compararlas en el gráfico",
    },
    col: { source: "Fuente", pair: "Par", bid: "Compra (bid)", ask: "Venta (ask)", rate: "Tipo", spread: "Diferencial", fee: "Comisión", effBuy: "Compra con comisión", effSell: "Venta con comisión", change: "24h", updated: "Actualizado", status: "Estado", chart: "Gráfico" },
    state: { ok: "OK", error: "Error", waiting: "Esperando", excluded: "Excluida", na: "Sin cotización" },
    type: { exchange: "Exchange", reference: "Tipo de referencia" },
    markets: { title: "Matriz de mercados", hint: "Precios en vivo en cada exchange. Verde — más barato para comprar, azul — mejor para vender (con comisiones).", compare: "Comparación de rendimiento", compareHint: "Cambio en % durante el periodo elegido", fiat: "Tipos fiduciarios por fuente" },
    alerts: {
      title: "Alertas de precio", new: "Nueva alerta", edit: "Editar alerta", empty: "Aún no hay alertas. Crea una para recibir avisos cuando el precio se mueva.",
      history: "Historial de notificaciones", clear: "Borrar", noEvents: "Aún no hay notificaciones", asset: "Moneda", vs: "Cotizada en",
      source: "Fuente", condition: "Condición", value: "Valor", window: "Periodo", repeat: "Repetir tras cada aviso",
      note: "Nota", notePh: "Comentario opcional", save: "Guardar", cancel: "Cancelar", delete: "Eliminar", deleteConfirm: "¿Eliminar esta alerta?",
      enable: "Activar", disable: "Pausar", active: "Activa", paused: "En pausa", fired: "Disparada {n} veces", now: "Ahora {v}",
      current: "Precio actual: {v}", fireAt: "Se dispara en {v}", baseline: "Referencia: {v}",
    },
    cond: { rise_pct: "Sube un %", fall_pct: "Baja un %", move_pct: "Varía ±%", rise_abs: "Sube una cantidad", fall_abs: "Baja una cantidad", above: "Supera", below: "Cae por debajo de" },
    win: { since: "Desde que se creó la alerta", "5m": "En 5 minutos", "15m": "En 15 minutos", "1h": "En 1 hora", "4h": "En 4 horas", "24h": "En 24 horas", "7d": "En 7 días" },
    msg: { up: "{asset} subió {chg} {window}", down: "{asset} bajó {chg} {window}", move: "{asset} cambió {chg} {window}", above: "{asset} superó {level}", below: "{asset} cayó por debajo de {level}", since: "desde que se creó la alerta", in: "en {w}" },
    unit: { min: "min", h: "h", d: "d" },
    settings: {
      title: "Ajustes", appearance: "Apariencia", language: "Idioma", theme: "Tema", scale: "Escala de la interfaz",
      scaleAuto: "Automática (según tamaño de ventana y resolución)", chartStyle: "Estilo del gráfico", currency: "Moneda de visualización",
      notifications: "Notificaciones", osNotify: "Notificaciones de escritorio",
      osNotifyHint: "Notificaciones nativas de Windows, macOS o Linux: funcionan incluso con esta ventana cerrada",
      browserNotify: "Notificaciones del navegador", sound: "Sonido", test: "Enviar notificación de prueba", testSent: "Notificación de prueba enviada",
      permDenied: "El navegador ha bloqueado las notificaciones", sources: "Exchanges y comisiones",
      sourcesHint: "Las fuentes sin marcar no se tienen en cuenta al buscar el mejor tipo. Las comisiones son de taker y se usan para el precio efectivo.",
      include: "Usar", reset: "Restablecer", services: "Microservicios", about: "Acerca de",
      aboutText: "Gratuito y de código abierto. Los datos provienen de API públicas de exchanges y bancos centrales, sin claves ni servicios de pago.",
      quit: "Salir de Cryptomon", quitConfirm: "¿Detener todos los servicios y salir?", stopped: "Cryptomon se ha detenido. Puedes cerrar esta ventana.",
    },
    theme: { system: "Sistema", light: "Claro", dark: "Oscuro", midnight: "Medianoche", ocean: "Océano", sepia: "Sepia" },
    svc: { up: "En marcha", down: "No disponible", waiting: "Iniciando" },
    toast: { saved: "Guardado", error: "Error: {e}" },
    asset: { BTC: "Bitcoin", ETH: "Ethereum", USDT: "Tether", TON: "Toncoin", USD: "Dólar estadounidense", EUR: "Euro", RUB: "Rublo ruso", CNY: "Yuan chino", JPY: "Yen japonés", INR: "Rupia india", CHF: "Franco suizo", GBP: "Libra esterlina" },
  },

  fr: {
    nav: { dashboard: "Tableau de bord", markets: "Marchés", alerts: "Alertes", settings: "Paramètres" },
    conn: { live: "En direct", offline: "Hors ligne", connecting: "Connexion…" },
    section: { crypto: "Cryptomonnaies", fiat: "Devises fiduciaires" },
    dash: { subtitle: "Cours en temps réel de {n} plateformes et sources", arbitrage: "Meilleur arbitrage actuel", services: "Microservices en ligne", alerts: "Alertes actives" },
    card: { buy: "Acheter", sell: "Vendre", sources: "{n} sur {m} sources", waiting: "En attente des données…", down: "Service indisponible", vsNote: "en {vs}" },
    time: { now: "à l’instant", sec: "il y a {n} s", min: "il y a {n} min", hour: "il y a {n} h", day: "il y a {n} j" },
    range: { "1h": "1H", "6h": "6H", "24h": "24H", "7d": "7J", "30d": "30J", "90d": "90J", "1y": "1A", "5y": "5A" },
    detail: {
      back: "Retour", aggregate: "Cours agrégé", percent: "Variation %", area: "Aire", line: "Ligne",
      best: "Meilleurs cours", bestBuy: "Où acheter au mieux", bestSell: "Où vendre au mieux", arbitrage: "Arbitrage après frais",
      arbHint: "Acheter sur {buy}, vendre sur {sell}", noArb: "Aucun arbitrage rentable pour l’instant", saving: "Économie",
      savingHint: "Meilleur achat comparé à la source la plus chère", converter: "Calculateur de change", amount: "Montant",
      buy: "Acheter", sell: "Vendre", cost: "Vous payez", receive: "Vous recevez", bestAt: "Meilleur : {src}", worstAt: "Pire : {src}",
      diff: "Différence", sources: "Plateformes et sources", createAlert: "Créer une alerte", noHistory: "Pas encore d’historique pour cette période",
      loading: "Chargement…", chartHint: "Cliquez sur les sources pour les comparer sur le graphique",
    },
    col: { source: "Source", pair: "Paire", bid: "Achat (bid)", ask: "Vente (ask)", rate: "Cours", spread: "Écart", fee: "Frais", effBuy: "Achat frais inclus", effSell: "Vente frais inclus", change: "24 h", updated: "Mis à jour", status: "Statut", chart: "Graphique" },
    state: { ok: "OK", error: "Erreur", waiting: "En attente", excluded: "Exclue", na: "Non coté" },
    type: { exchange: "Plateforme", reference: "Cours de référence" },
    markets: { title: "Matrice des marchés", hint: "Cours en direct sur chaque plateforme. Vert — achat le moins cher, bleu — meilleure vente (frais inclus).", compare: "Comparaison des performances", compareHint: "Variation en % sur la période choisie", fiat: "Cours des devises par source" },
    alerts: {
      title: "Alertes de prix", new: "Nouvelle alerte", edit: "Modifier l’alerte", empty: "Aucune alerte pour l’instant. Créez-en une pour être averti lorsqu’un cours bouge.",
      history: "Historique des notifications", clear: "Effacer", noEvents: "Aucune notification pour l’instant", asset: "Devise", vs: "Exprimée en",
      source: "Source", condition: "Condition", value: "Valeur", window: "Période", repeat: "Répéter après chaque notification",
      note: "Note", notePh: "Commentaire facultatif", save: "Enregistrer", cancel: "Annuler", delete: "Supprimer", deleteConfirm: "Supprimer cette alerte ?",
      enable: "Activer", disable: "Suspendre", active: "Active", paused: "Suspendue", fired: "Déclenchée {n} fois", now: "Actuel {v}",
      current: "Cours actuel : {v}", fireAt: "Se déclenche à {v}", baseline: "Référence : {v}",
    },
    cond: { rise_pct: "Monte de %", fall_pct: "Baisse de %", move_pct: "Varie de ±%", rise_abs: "Monte d’un montant", fall_abs: "Baisse d’un montant", above: "Passe au-dessus de", below: "Passe sous" },
    win: { since: "Depuis la création de l’alerte", "5m": "En 5 minutes", "15m": "En 15 minutes", "1h": "En 1 heure", "4h": "En 4 heures", "24h": "En 24 heures", "7d": "En 7 jours" },
    msg: { up: "{asset} a augmenté de {chg} {window}", down: "{asset} a baissé de {chg} {window}", move: "{asset} a varié de {chg} {window}", above: "{asset} est passé au-dessus de {level}", below: "{asset} est passé sous {level}", since: "depuis la création de l’alerte", in: "en {w}" },
    unit: { min: "min", h: "h", d: "j" },
    settings: {
      title: "Paramètres", appearance: "Apparence", language: "Langue", theme: "Thème", scale: "Échelle de l’interface",
      scaleAuto: "Automatique (selon la taille de la fenêtre et la résolution)", chartStyle: "Style du graphique", currency: "Devise d’affichage",
      notifications: "Notifications", osNotify: "Notifications du bureau",
      osNotifyHint: "Notifications natives de Windows, macOS ou Linux — elles fonctionnent même fenêtre fermée",
      browserNotify: "Notifications du navigateur", sound: "Son", test: "Envoyer une notification de test", testSent: "Notification de test envoyée",
      permDenied: "Le navigateur bloque les notifications", sources: "Plateformes et frais",
      sourcesHint: "Les sources décochées sont ignorées dans la recherche du meilleur cours. Les frais sont des frais taker utilisés pour le prix effectif.",
      include: "Utiliser", reset: "Réinitialiser", services: "Microservices", about: "À propos",
      aboutText: "Gratuit et open source. Données issues des API publiques des plateformes et des banques centrales, sans clé ni service payant.",
      quit: "Quitter Cryptomon", quitConfirm: "Arrêter tous les services et quitter ?", stopped: "Cryptomon est arrêté. Vous pouvez fermer cette fenêtre.",
    },
    theme: { system: "Système", light: "Clair", dark: "Sombre", midnight: "Minuit", ocean: "Océan", sepia: "Sépia" },
    svc: { up: "En marche", down: "Indisponible", waiting: "Démarrage" },
    toast: { saved: "Enregistré", error: "Erreur : {e}" },
    asset: { BTC: "Bitcoin", ETH: "Ethereum", USDT: "Tether", TON: "Toncoin", USD: "Dollar américain", EUR: "Euro", RUB: "Rouble russe", CNY: "Yuan chinois", JPY: "Yen japonais", INR: "Roupie indienne", CHF: "Franc suisse", GBP: "Livre sterling" },
  },

  ar: {
    nav: { dashboard: "لوحة المعلومات", markets: "الأسواق", alerts: "التنبيهات", settings: "الإعدادات" },
    conn: { live: "مباشر", offline: "غير متصل", connecting: "جارٍ الاتصال…" },
    section: { crypto: "العملات المشفرة", fiat: "العملات الورقية" },
    dash: { subtitle: "أسعار فورية من {n} منصة ومصدر", arbitrage: "أفضل فرصة مراجحة الآن", services: "الخدمات المصغّرة المتصلة", alerts: "التنبيهات النشطة" },
    card: { buy: "شراء", sell: "بيع", sources: "{n} من {m} مصادر", waiting: "في انتظار البيانات…", down: "الخدمة غير متاحة", vsNote: "بـ {vs}" },
    time: { now: "الآن", sec: "قبل {n} ث", min: "قبل {n} د", hour: "قبل {n} س", day: "قبل {n} يوم" },
    range: { "1h": "1 س", "6h": "6 س", "24h": "24 س", "7d": "7 أيام", "30d": "30 يومًا", "90d": "90 يومًا", "1y": "سنة", "5y": "5 سنوات" },
    detail: {
      back: "رجوع", aggregate: "السعر المجمّع", percent: "التغيّر %", area: "مساحة", line: "خط",
      best: "أفضل الأسعار", bestBuy: "أفضل مكان للشراء", bestSell: "أفضل مكان للبيع", arbitrage: "المراجحة بعد الرسوم",
      arbHint: "اشترِ من {buy} وبِع في {sell}", noArb: "لا توجد فرصة مراجحة مربحة حاليًا", saving: "التوفير",
      savingHint: "أفضل سعر شراء مقارنةً بأغلى مصدر", converter: "حاسبة الصرف", amount: "الكمية",
      buy: "شراء", sell: "بيع", cost: "ستدفع", receive: "ستحصل على", bestAt: "الأفضل: {src}", worstAt: "الأسوأ: {src}",
      diff: "الفرق", sources: "المنصات والمصادر", createAlert: "إنشاء تنبيه", noHistory: "لا يوجد سجل لهذه الفترة بعد",
      loading: "جارٍ التحميل…", chartHint: "انقر على المصادر لمقارنتها على الرسم البياني",
    },
    col: { source: "المصدر", pair: "الزوج", bid: "سعر الطلب", ask: "سعر العرض", rate: "السعر", spread: "الفارق", fee: "الرسوم", effBuy: "الشراء مع الرسوم", effSell: "البيع مع الرسوم", change: "24 س", updated: "آخر تحديث", status: "الحالة", chart: "الرسم" },
    state: { ok: "سليم", error: "خطأ", waiting: "انتظار", excluded: "مستبعد", na: "غير مُسعَّر" },
    type: { exchange: "منصة تداول", reference: "سعر مرجعي" },
    markets: { title: "مصفوفة الأسواق", hint: "أسعار مباشرة في كل منصة. الأخضر — الأرخص للشراء، الأزرق — الأفضل للبيع (شاملة الرسوم).", compare: "مقارنة الأداء", compareHint: "التغيّر بالنسبة المئوية خلال الفترة المحددة", fiat: "أسعار العملات الورقية حسب المصدر" },
    alerts: {
      title: "تنبيهات الأسعار", new: "تنبيه جديد", edit: "تعديل التنبيه", empty: "لا توجد تنبيهات بعد. أنشئ تنبيهًا لتعرف عند تحرّك السعر.",
      history: "سجل الإشعارات", clear: "مسح", noEvents: "لا توجد إشعارات بعد", asset: "العملة", vs: "مُسعّرة بـ",
      source: "المصدر", condition: "الشرط", value: "القيمة", window: "الفترة", repeat: "التكرار بعد كل إشعار",
      note: "ملاحظة", notePh: "تعليق اختياري", save: "حفظ", cancel: "إلغاء", delete: "حذف", deleteConfirm: "حذف هذا التنبيه؟",
      enable: "تفعيل", disable: "إيقاف مؤقت", active: "نشط", paused: "متوقف", fired: "انطلق {n} مرة", now: "الآن {v}",
      current: "السعر الحالي: {v}", fireAt: "ينطلق عند {v}", baseline: "المرجع: {v}",
    },
    cond: { rise_pct: "يرتفع بنسبة %", fall_pct: "ينخفض بنسبة %", move_pct: "يتغيّر بنسبة ±%", rise_abs: "يرتفع بمقدار", fall_abs: "ينخفض بمقدار", above: "يتجاوز", below: "ينزل دون" },
    win: { since: "منذ إنشاء التنبيه", "5m": "خلال 5 دقائق", "15m": "خلال 15 دقيقة", "1h": "خلال ساعة", "4h": "خلال 4 ساعات", "24h": "خلال 24 ساعة", "7d": "خلال 7 أيام" },
    msg: { up: "ارتفع {asset} بمقدار {chg} {window}", down: "انخفض {asset} بمقدار {chg} {window}", move: "تغيّر {asset} بمقدار {chg} {window}", above: "ارتفع {asset} فوق {level}", below: "انخفض {asset} دون {level}", since: "منذ إنشاء التنبيه", in: "خلال {w}" },
    unit: { min: "دقيقة", h: "ساعة", d: "يوم" },
    settings: {
      title: "الإعدادات", appearance: "المظهر", language: "اللغة", theme: "السمة", scale: "حجم الواجهة",
      scaleAuto: "تلقائي (حسب حجم النافذة والدقة)", chartStyle: "نمط الرسم البياني", currency: "عملة العرض",
      notifications: "الإشعارات", osNotify: "إشعارات سطح المكتب",
      osNotifyHint: "إشعارات أصلية لنظام Windows أو macOS أو Linux — تعمل حتى عند إغلاق هذه النافذة",
      browserNotify: "إشعارات المتصفح", sound: "الصوت", test: "إرسال إشعار تجريبي", testSent: "تم إرسال الإشعار التجريبي",
      permDenied: "المتصفح يحظر الإشعارات", sources: "المنصات والرسوم",
      sourcesHint: "المصادر غير المحددة لا تُحتسب عند البحث عن أفضل سعر. الرسوم هي رسوم المتلقّي وتُستخدم لحساب السعر الفعلي.",
      include: "استخدام", reset: "إعادة تعيين", services: "الخدمات المصغّرة", about: "حول",
      aboutText: "مجاني ومفتوح المصدر. البيانات من واجهات برمجة عامة للمنصات والبنوك المركزية — دون مفاتيح أو خدمات مدفوعة.",
      quit: "إنهاء Cryptomon", quitConfirm: "إيقاف جميع الخدمات والخروج؟", stopped: "تم إيقاف Cryptomon. يمكنك إغلاق هذه النافذة.",
    },
    theme: { system: "النظام", light: "فاتح", dark: "داكن", midnight: "منتصف الليل", ocean: "المحيط", sepia: "سيبيا" },
    svc: { up: "يعمل", down: "غير متاح", waiting: "قيد التشغيل" },
    toast: { saved: "تم الحفظ", error: "خطأ: {e}" },
    asset: { BTC: "بيتكوين", ETH: "إيثريوم", USDT: "تيثر", TON: "تونكوين", USD: "الدولار الأمريكي", EUR: "اليورو", RUB: "الروبل الروسي", CNY: "اليوان الصيني", JPY: "الين الياباني", INR: "الروبية الهندية", CHF: "الفرنك السويسري", GBP: "الجنيه الإسترليني" },
  },

  bn: {
    nav: { dashboard: "ড্যাশবোর্ড", markets: "বাজার", alerts: "সতর্কতা", settings: "সেটিংস" },
    conn: { live: "লাইভ", offline: "অফলাইন", connecting: "সংযোগ হচ্ছে…" },
    section: { crypto: "ক্রিপ্টোকারেন্সি", fiat: "ফিয়াট মুদ্রা" },
    dash: { subtitle: "{n}টি এক্সচেঞ্জ ও উৎস থেকে রিয়েল-টাইম দাম", arbitrage: "এখনকার সেরা আরবিট্রাজ", services: "অনলাইন মাইক্রোসার্ভিস", alerts: "সক্রিয় সতর্কতা" },
    card: { buy: "কিনুন", sell: "বিক্রি করুন", sources: "{m}টির মধ্যে {n}টি উৎস", waiting: "ডেটার অপেক্ষায়…", down: "পরিষেবা অনুপলব্ধ", vsNote: "{vs}-এ" },
    time: { now: "এইমাত্র", sec: "{n} সেকেন্ড আগে", min: "{n} মিনিট আগে", hour: "{n} ঘণ্টা আগে", day: "{n} দিন আগে" },
    range: { "1h": "1ঘ", "6h": "6ঘ", "24h": "24ঘ", "7d": "7দি", "30d": "30দি", "90d": "90দি", "1y": "1ব", "5y": "5ব" },
    detail: {
      back: "ফিরে যান", aggregate: "সমন্বিত দাম", percent: "% পরিবর্তন", area: "এরিয়া", line: "লাইন",
      best: "সেরা হার", bestBuy: "কেনার সেরা জায়গা", bestSell: "বিক্রির সেরা জায়গা", arbitrage: "ফি-র পরে আরবিট্রাজ",
      arbHint: "{buy}-এ কিনুন, {sell}-এ বিক্রি করুন", noArb: "এখন কোনো লাভজনক আরবিট্রাজ নেই", saving: "সাশ্রয়",
      savingHint: "সেরা ক্রয় বনাম সবচেয়ে দামি উৎস", converter: "বিনিময় ক্যালকুলেটর", amount: "পরিমাণ",
      buy: "কিনুন", sell: "বিক্রি করুন", cost: "আপনি দেবেন", receive: "আপনি পাবেন", bestAt: "সেরা: {src}", worstAt: "সবচেয়ে খারাপ: {src}",
      diff: "পার্থক্য", sources: "এক্সচেঞ্জ ও উৎস", createAlert: "সতর্কতা তৈরি করুন", noHistory: "এই সময়ের ইতিহাস এখনো নেই",
      loading: "লোড হচ্ছে…", chartHint: "চার্টে তুলনা করতে উৎসগুলিতে ক্লিক করুন",
    },
    col: { source: "উৎস", pair: "জোড়া", bid: "বিড", ask: "আস্ক", rate: "হার", spread: "স্প্রেড", fee: "ফি", effBuy: "ফি-সহ ক্রয়", effSell: "ফি-সহ বিক্রয়", change: "24ঘ", updated: "হালনাগাদ", status: "অবস্থা", chart: "চার্ট" },
    state: { ok: "ঠিক আছে", error: "ত্রুটি", waiting: "অপেক্ষমাণ", excluded: "বাদ দেওয়া", na: "দর নেই" },
    type: { exchange: "এক্সচেঞ্জ", reference: "রেফারেন্স হার" },
    markets: { title: "বাজার ম্যাট্রিক্স", hint: "প্রতিটি এক্সচেঞ্জে লাইভ দাম। সবুজ — কেনার জন্য সবচেয়ে সস্তা, নীল — বিক্রির জন্য সেরা (ফি-সহ)।", compare: "পারফরম্যান্স তুলনা", compareHint: "নির্বাচিত সময়ে % পরিবর্তন", fiat: "উৎস অনুযায়ী ফিয়াট হার" },
    alerts: {
      title: "দামের সতর্কতা", new: "নতুন সতর্কতা", edit: "সতর্কতা সম্পাদনা", empty: "এখনো কোনো সতর্কতা নেই। দাম নড়লে জানতে একটি তৈরি করুন।",
      history: "বিজ্ঞপ্তির ইতিহাস", clear: "মুছুন", noEvents: "এখনো কোনো বিজ্ঞপ্তি নেই", asset: "মুদ্রা", vs: "মূল্য মুদ্রা",
      source: "উৎস", condition: "শর্ত", value: "মান", window: "সময়কাল", repeat: "প্রতিটি বিজ্ঞপ্তির পরে পুনরাবৃত্তি",
      note: "নোট", notePh: "ঐচ্ছিক মন্তব্য", save: "সংরক্ষণ", cancel: "বাতিল", delete: "মুছুন", deleteConfirm: "এই সতর্কতাটি মুছবেন?",
      enable: "চালু করুন", disable: "বিরতি", active: "সক্রিয়", paused: "বিরতিতে", fired: "{n} বার ট্রিগার হয়েছে", now: "এখন {v}",
      current: "বর্তমান দাম: {v}", fireAt: "{v}-এ ট্রিগার হবে", baseline: "রেফারেন্স: {v}",
    },
    cond: { rise_pct: "% বাড়লে", fall_pct: "% কমলে", move_pct: "±% বদলালে", rise_abs: "পরিমাণে বাড়লে", fall_abs: "পরিমাণে কমলে", above: "এর উপরে গেলে", below: "এর নিচে নামলে" },
    win: { since: "সতর্কতা তৈরির পর থেকে", "5m": "5 মিনিটে", "15m": "15 মিনিটে", "1h": "1 ঘণ্টায়", "4h": "4 ঘণ্টায়", "24h": "24 ঘণ্টায়", "7d": "7 দিনে" },
    msg: { up: "{asset} {window} {chg} বেড়েছে", down: "{asset} {window} {chg} কমেছে", move: "{asset} {window} {chg} পরিবর্তিত হয়েছে", above: "{asset} {level}-এর উপরে উঠেছে", below: "{asset} {level}-এর নিচে নেমেছে", since: "অ্যালার্ট সেট করার পর থেকে", in: "{w}-এ" },
    unit: { min: "মিনিট", h: "ঘণ্টা", d: "দিন" },
    settings: {
      title: "সেটিংস", appearance: "চেহারা", language: "ভাষা", theme: "থিম", scale: "ইন্টারফেস স্কেল",
      scaleAuto: "স্বয়ংক্রিয় (উইন্ডোর আকার ও রেজোলিউশন অনুযায়ী)", chartStyle: "চার্টের ধরন", currency: "প্রদর্শন মুদ্রা",
      notifications: "বিজ্ঞপ্তি", osNotify: "ডেস্কটপ বিজ্ঞপ্তি",
      osNotifyHint: "Windows, macOS বা Linux-এর নিজস্ব বিজ্ঞপ্তি — এই উইন্ডো বন্ধ থাকলেও কাজ করে",
      browserNotify: "ব্রাউজার বিজ্ঞপ্তি", sound: "শব্দ", test: "পরীক্ষামূলক বিজ্ঞপ্তি পাঠান", testSent: "পরীক্ষামূলক বিজ্ঞপ্তি পাঠানো হয়েছে",
      permDenied: "ব্রাউজার বিজ্ঞপ্তি অবরুদ্ধ করেছে", sources: "এক্সচেঞ্জ ও ফি",
      sourcesHint: "টিক না দেওয়া উৎসগুলি সেরা হার খোঁজার সময় বিবেচিত হয় না। ফি হলো টেকার ফি, কার্যকর দাম হিসাবে ব্যবহৃত।",
      include: "ব্যবহার", reset: "রিসেট", services: "মাইক্রোসার্ভিস", about: "সম্পর্কে",
      aboutText: "বিনামূল্যে ও ওপেন সোর্স। ডেটা আসে এক্সচেঞ্জের পাবলিক API ও কেন্দ্রীয় ব্যাংক থেকে — কোনো কী বা পেইড সার্ভিস লাগে না।",
      quit: "Cryptomon বন্ধ করুন", quitConfirm: "সব পরিষেবা থামিয়ে বেরিয়ে যাবেন?", stopped: "Cryptomon বন্ধ হয়েছে। আপনি এই উইন্ডো বন্ধ করতে পারেন।",
    },
    theme: { system: "সিস্টেম", light: "হালকা", dark: "গাঢ়", midnight: "মধ্যরাত", ocean: "সমুদ্র", sepia: "সেপিয়া" },
    svc: { up: "চলছে", down: "অনুপলব্ধ", waiting: "চালু হচ্ছে" },
    toast: { saved: "সংরক্ষিত", error: "ত্রুটি: {e}" },
    asset: { BTC: "বিটকয়েন", ETH: "ইথেরিয়াম", USDT: "টেথার", TON: "টনকয়েন", USD: "মার্কিন ডলার", EUR: "ইউরো", RUB: "রুশ রুবল", CNY: "চীনা ইউয়ান", JPY: "জাপানি ইয়েন", INR: "ভারতীয় রুপি", CHF: "সুইস ফ্রাঁ", GBP: "ব্রিটিশ পাউন্ড" },
  },

  pt: {
    nav: { dashboard: "Painel", markets: "Mercados", alerts: "Alertas", settings: "Configurações" },
    conn: { live: "Ao vivo", offline: "Offline", connecting: "Conectando…" },
    section: { crypto: "Criptomoedas", fiat: "Moedas fiduciárias" },
    dash: { subtitle: "Preços em tempo real de {n} corretoras e fontes", arbitrage: "Melhor arbitragem agora", services: "Microsserviços online", alerts: "Alertas ativos" },
    card: { buy: "Comprar", sell: "Vender", sources: "{n} de {m} fontes", waiting: "Aguardando dados…", down: "Serviço indisponível", vsNote: "em {vs}" },
    time: { now: "agora mesmo", sec: "há {n} s", min: "há {n} min", hour: "há {n} h", day: "há {n} d" },
    range: { "1h": "1H", "6h": "6H", "24h": "24H", "7d": "7D", "30d": "30D", "90d": "90D", "1y": "1A", "5y": "5A" },
    detail: {
      back: "Voltar", aggregate: "Preço agregado", percent: "Variação %", area: "Área", line: "Linha",
      best: "Melhores cotações", bestBuy: "Onde comprar melhor", bestSell: "Onde vender melhor", arbitrage: "Arbitragem após taxas",
      arbHint: "Comprar na {buy}, vender na {sell}", noArb: "Sem arbitragem lucrativa no momento", saving: "Economia",
      savingHint: "Melhor compra vs. a fonte mais cara", converter: "Calculadora de câmbio", amount: "Quantidade",
      buy: "Comprar", sell: "Vender", cost: "Você paga", receive: "Você recebe", bestAt: "Melhor: {src}", worstAt: "Pior: {src}",
      diff: "Diferença", sources: "Corretoras e fontes", createAlert: "Criar alerta", noHistory: "Ainda não há histórico para este período",
      loading: "Carregando…", chartHint: "Clique nas fontes para compará-las no gráfico",
    },
    col: { source: "Fonte", pair: "Par", bid: "Compra (bid)", ask: "Venda (ask)", rate: "Cotação", spread: "Spread", fee: "Taxa", effBuy: "Compra c/ taxa", effSell: "Venda c/ taxa", change: "24h", updated: "Atualizado", status: "Status", chart: "Gráfico" },
    state: { ok: "OK", error: "Erro", waiting: "Aguardando", excluded: "Excluída", na: "Sem cotação" },
    type: { exchange: "Corretora", reference: "Cotação de referência" },
    markets: { title: "Matriz de mercados", hint: "Preços ao vivo em cada corretora. Verde — mais barato para comprar, azul — melhor para vender (com taxas).", compare: "Comparação de desempenho", compareHint: "Variação em % no período selecionado", fiat: "Cotações fiduciárias por fonte" },
    alerts: {
      title: "Alertas de preço", new: "Novo alerta", edit: "Editar alerta", empty: "Ainda não há alertas. Crie um para ser avisado quando o preço mudar.",
      history: "Histórico de notificações", clear: "Limpar", noEvents: "Ainda não há notificações", asset: "Moeda", vs: "Cotada em",
      source: "Fonte", condition: "Condição", value: "Valor", window: "Período", repeat: "Repetir após cada notificação",
      note: "Nota", notePh: "Comentário opcional", save: "Salvar", cancel: "Cancelar", delete: "Excluir", deleteConfirm: "Excluir este alerta?",
      enable: "Ativar", disable: "Pausar", active: "Ativo", paused: "Pausado", fired: "Disparado {n}×", now: "Agora {v}",
      current: "Preço atual: {v}", fireAt: "Dispara em {v}", baseline: "Referência: {v}",
    },
    cond: { rise_pct: "Sobe %", fall_pct: "Cai %", move_pct: "Varia ±%", rise_abs: "Sobe um valor", fall_abs: "Cai um valor", above: "Fica acima de", below: "Fica abaixo de" },
    win: { since: "Desde a criação do alerta", "5m": "Em 5 minutos", "15m": "Em 15 minutos", "1h": "Em 1 hora", "4h": "Em 4 horas", "24h": "Em 24 horas", "7d": "Em 7 dias" },
    msg: { up: "{asset} subiu {chg} {window}", down: "{asset} caiu {chg} {window}", move: "{asset} variou {chg} {window}", above: "{asset} subiu acima de {level}", below: "{asset} caiu abaixo de {level}", since: "desde a criação do alerta", in: "em {w}" },
    unit: { min: "min", h: "h", d: "d" },
    settings: {
      title: "Configurações", appearance: "Aparência", language: "Idioma", theme: "Tema", scale: "Escala da interface",
      scaleAuto: "Automática (pelo tamanho da janela e resolução)", chartStyle: "Estilo do gráfico", currency: "Moeda de exibição",
      notifications: "Notificações", osNotify: "Notificações da área de trabalho",
      osNotifyHint: "Notificações nativas do Windows, macOS ou Linux — funcionam mesmo com esta janela fechada",
      browserNotify: "Notificações do navegador", sound: "Som", test: "Enviar notificação de teste", testSent: "Notificação de teste enviada",
      permDenied: "O navegador bloqueou as notificações", sources: "Corretoras e taxas",
      sourcesHint: "Fontes desmarcadas são ignoradas na busca pela melhor cotação. As taxas são de taker e servem para o preço efetivo.",
      include: "Usar", reset: "Redefinir", services: "Microsserviços", about: "Sobre",
      aboutText: "Gratuito e de código aberto. Os dados vêm de APIs públicas de corretoras e bancos centrais — sem chaves nem serviços pagos.",
      quit: "Sair do Cryptomon", quitConfirm: "Parar todos os serviços e sair?", stopped: "O Cryptomon foi encerrado. Você pode fechar esta janela.",
    },
    theme: { system: "Sistema", light: "Claro", dark: "Escuro", midnight: "Meia-noite", ocean: "Oceano", sepia: "Sépia" },
    svc: { up: "Em execução", down: "Indisponível", waiting: "Iniciando" },
    toast: { saved: "Salvo", error: "Erro: {e}" },
    asset: { BTC: "Bitcoin", ETH: "Ethereum", USDT: "Tether", TON: "Toncoin", USD: "Dólar americano", EUR: "Euro", RUB: "Rublo russo", CNY: "Yuan chinês", JPY: "Iene japonês", INR: "Rupia indiana", CHF: "Franco suíço", GBP: "Libra esterlina" },
  },

  ur: {
    nav: { dashboard: "ڈیش بورڈ", markets: "مارکیٹس", alerts: "الرٹس", settings: "ترتیبات" },
    conn: { live: "لائیو", offline: "آف لائن", connecting: "جڑ رہا ہے…" },
    section: { crypto: "کرپٹو کرنسیاں", fiat: "فیاٹ کرنسیاں" },
    dash: { subtitle: "{n} ایکسچینجز اور ذرائع سے ریئل ٹائم قیمتیں", arbitrage: "ابھی بہترین آربٹریج", services: "آن لائن مائیکرو سروسز", alerts: "فعال الرٹس" },
    card: { buy: "خریدیں", sell: "بیچیں", sources: "{m} میں سے {n} ذرائع", waiting: "ڈیٹا کا انتظار…", down: "سروس دستیاب نہیں", vsNote: "{vs} میں" },
    time: { now: "ابھی ابھی", sec: "{n} سیکنڈ پہلے", min: "{n} منٹ پہلے", hour: "{n} گھنٹے پہلے", day: "{n} دن پہلے" },
    range: { "1h": "1 گھنٹہ", "6h": "6 گھنٹے", "24h": "24 گھنٹے", "7d": "7 دن", "30d": "30 دن", "90d": "90 دن", "1y": "1 سال", "5y": "5 سال" },
    detail: {
      back: "واپس", aggregate: "مجموعی قیمت", percent: "% تبدیلی", area: "ایریا", line: "لائن",
      best: "بہترین ریٹس", bestBuy: "خریدنے کی بہترین جگہ", bestSell: "بیچنے کی بہترین جگہ", arbitrage: "فیس کے بعد آربٹریج",
      arbHint: "{buy} پر خریدیں، {sell} پر بیچیں", noArb: "ابھی کوئی منافع بخش آربٹریج نہیں", saving: "بچت",
      savingHint: "بہترین خرید بمقابلہ سب سے مہنگا ذریعہ", converter: "تبادلہ کیلکولیٹر", amount: "مقدار",
      buy: "خریدیں", sell: "بیچیں", cost: "آپ ادا کریں گے", receive: "آپ کو ملے گا", bestAt: "بہترین: {src}", worstAt: "بدترین: {src}",
      diff: "فرق", sources: "ایکسچینجز اور ذرائع", createAlert: "الرٹ بنائیں", noHistory: "اس مدت کی تاریخ ابھی دستیاب نہیں",
      loading: "لوڈ ہو رہا ہے…", chartHint: "چارٹ پر موازنہ کے لیے ذرائع پر کلک کریں",
    },
    col: { source: "ذریعہ", pair: "جوڑا", bid: "بِڈ", ask: "آسک", rate: "ریٹ", spread: "اسپریڈ", fee: "فیس", effBuy: "فیس سمیت خرید", effSell: "فیس سمیت فروخت", change: "24 گھنٹے", updated: "اپ ڈیٹ", status: "حالت", chart: "چارٹ" },
    state: { ok: "ٹھیک", error: "خرابی", waiting: "انتظار", excluded: "خارج", na: "قیمت دستیاب نہیں" },
    type: { exchange: "ایکسچینج", reference: "حوالہ جاتی ریٹ" },
    markets: { title: "مارکیٹ میٹرکس", hint: "ہر ایکسچینج پر لائیو قیمتیں۔ سبز — خریدنے کے لیے سب سے سستا، نیلا — بیچنے کے لیے بہترین (فیس سمیت)۔", compare: "کارکردگی کا موازنہ", compareHint: "منتخب مدت میں % تبدیلی", fiat: "ذرائع کے لحاظ سے فیاٹ ریٹس" },
    alerts: {
      title: "قیمت کے الرٹس", new: "نیا الرٹ", edit: "الرٹ میں ترمیم", empty: "ابھی کوئی الرٹ نہیں۔ قیمت بدلنے پر اطلاع کے لیے ایک الرٹ بنائیں۔",
      history: "اطلاعات کی تاریخ", clear: "صاف کریں", noEvents: "ابھی کوئی اطلاع نہیں", asset: "کرنسی", vs: "قیمت کی کرنسی",
      source: "ذریعہ", condition: "شرط", value: "قدر", window: "مدت", repeat: "ہر اطلاع کے بعد دہرائیں",
      note: "نوٹ", notePh: "اختیاری تبصرہ", save: "محفوظ کریں", cancel: "منسوخ", delete: "حذف کریں", deleteConfirm: "یہ الرٹ حذف کریں؟",
      enable: "فعال کریں", disable: "روکیں", active: "فعال", paused: "رکا ہوا", fired: "{n} بار متحرک ہوا", now: "ابھی {v}",
      current: "موجودہ قیمت: {v}", fireAt: "{v} پر متحرک ہوگا", baseline: "حوالہ: {v}",
    },
    cond: { rise_pct: "% بڑھے", fall_pct: "% گھٹے", move_pct: "±% بدلے", rise_abs: "رقم سے بڑھے", fall_abs: "رقم سے گھٹے", above: "اس سے اوپر جائے", below: "اس سے نیچے جائے" },
    win: { since: "الرٹ بنانے کے بعد سے", "5m": "5 منٹ میں", "15m": "15 منٹ میں", "1h": "1 گھنٹے میں", "4h": "4 گھنٹوں میں", "24h": "24 گھنٹوں میں", "7d": "7 دنوں میں" },
    msg: { up: "{asset} میں {window} {chg} اضافہ ہوا", down: "{asset} میں {window} {chg} کمی ہوئی", move: "{asset} میں {window} {chg} تبدیلی ہوئی", above: "{asset} {level} سے اوپر چلا گیا", below: "{asset} {level} سے نیچے آ گیا", since: "الرٹ سیٹ کرنے کے بعد سے", in: "{w} میں" },
    unit: { min: "منٹ", h: "گھنٹے", d: "دن" },
    settings: {
      title: "ترتیبات", appearance: "ظاہری شکل", language: "زبان", theme: "تھیم", scale: "انٹرفیس اسکیل",
      scaleAuto: "خودکار (ونڈو کے سائز اور ریزولوشن کے مطابق)", chartStyle: "چارٹ کا انداز", currency: "ڈسپلے کرنسی",
      notifications: "اطلاعات", osNotify: "ڈیسک ٹاپ اطلاعات",
      osNotifyHint: "Windows، macOS یا Linux کی مقامی اطلاعات — یہ ونڈو بند ہونے پر بھی کام کرتی ہیں",
      browserNotify: "براؤزر اطلاعات", sound: "آواز", test: "آزمائشی اطلاع بھیجیں", testSent: "آزمائشی اطلاع بھیج دی گئی",
      permDenied: "براؤزر نے اطلاعات بلاک کر دی ہیں", sources: "ایکسچینجز اور فیس",
      sourcesHint: "بغیر نشان والے ذرائع بہترین ریٹ کی تلاش میں شامل نہیں ہوتے۔ فیس ٹیکر فیس ہے جو مؤثر قیمت کے لیے استعمال ہوتی ہے۔",
      include: "استعمال", reset: "ری سیٹ", services: "مائیکرو سروسز", about: "تعارف",
      aboutText: "مفت اور اوپن سورس۔ ڈیٹا ایکسچینجز کے عوامی API اور مرکزی بینکوں سے — کسی کلید یا معاوضہ والی سروس کی ضرورت نہیں۔",
      quit: "Cryptomon بند کریں", quitConfirm: "تمام سروسز روک کر باہر نکلیں؟", stopped: "Cryptomon بند ہو گیا۔ آپ یہ ونڈو بند کر سکتے ہیں۔",
    },
    theme: { system: "سسٹم", light: "ہلکا", dark: "گہرا", midnight: "آدھی رات", ocean: "سمندر", sepia: "سیپیا" },
    svc: { up: "چل رہا ہے", down: "دستیاب نہیں", waiting: "شروع ہو رہا ہے" },
    toast: { saved: "محفوظ ہو گیا", error: "خرابی: {e}" },
    asset: { BTC: "بٹ کوائن", ETH: "ایتھیریم", USDT: "ٹیتھر", TON: "ٹون کوائن", USD: "امریکی ڈالر", EUR: "یورو", RUB: "روسی روبل", CNY: "چینی یوآن", JPY: "جاپانی ین", INR: "بھارتی روپیہ", CHF: "سوئس فرانک", GBP: "برطانوی پاؤنڈ" },
  },
};

function flatten(obj, prefix = "", out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? prefix + "." + k : k;
    if (v && typeof v === "object") flatten(v, key, out);
    else out[key] = v;
  }
  return out;
}

const FLAT = Object.fromEntries(Object.entries(D).map(([code, dict]) => [code, flatten(dict)]));

let current = "en";

export function detectLang() {
  const prefs = navigator.languages && navigator.languages.length ? navigator.languages : [navigator.language || "en"];
  for (const p of prefs) {
    const base = String(p).toLowerCase().split("-")[0];
    if (FLAT[base]) return base;
  }
  return "en";
}

export function setLang(code) {
  current = FLAT[code] ? code : "en";
  const info = LANGS.find((l) => l.code === current);
  document.documentElement.lang = current;
  document.documentElement.dir = info && info.rtl ? "rtl" : "ltr";
  return current;
}

export function getLang() {
  return current;
}

// Locale used for number/date formatting, always with Latin digits so
// that price tables stay aligned and comparable.
export function locale() {
  const map = { zh: "zh-CN", hi: "hi-IN", bn: "bn-BD", ar: "ar", ur: "ur-PK", pt: "pt-BR", es: "es-ES", fr: "fr-FR", ru: "ru-RU", en: "en-US" };
  return (map[current] || "en-US") + "-u-nu-latn";
}

export function t(key, vars) {
  let s = FLAT[current][key];
  if (s === undefined) s = FLAT.en[key];
  if (s === undefined) return key;
  if (vars) for (const [k, v] of Object.entries(vars)) s = s.split("{" + k + "}").join(String(v));
  return s;
}

// Used by the self-check in tests: keys missing per language.
export function missingKeys() {
  const ref = Object.keys(FLAT.en);
  const out = {};
  for (const code of Object.keys(FLAT)) {
    const miss = ref.filter((k) => FLAT[code][k] === undefined);
    if (miss.length) out[code] = miss;
  }
  return out;
}
