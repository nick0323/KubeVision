// Prettier 配置 - ES Module 版本
export default {
  // 每行最大字符数
  printWidth: 100,

  // 使用 Tab 缩进
  useTabs: false,

  // 缩进空格数
  tabWidth: 2,

  // 行尾使用分号
  semi: true,

  // 使用单引号
  singleQuote: true,

  // 对象属性引号
  quoteProps: 'as-needed',

  // JSX 使用双引号
  jsxSingleQuote: false,

  // 尾随逗号
  trailingComma: 'es5',

  // 对象字面量的大括号间添加空格
  bracketSpacing: true,

  // JSX 标签的 > 不另起一行
  bracketSameLine: false,

  // 箭头函数单个参数不加括号
  arrowParens: 'avoid',

  // 每个文件格式化的范围是整个文件
  rangeStart: 0,
  rangeEnd: Infinity,

  // 不需要 @prettier 注释即可格式化
  requirePragma: false,

  // 不在文件顶部插入 @format 注释
  insertPragma: false,

  // 在 markdown 文件中换行
  proseWrap: 'preserve',

  // HTML 空格敏感性
  htmlWhitespaceSensitivity: 'css',

  // 缩进样式表
  vueIndentScriptAndStyle: false,

  // 换行符
  endOfLine: 'lf',

  // 格式化嵌入的代码
  embeddedLanguageFormatting: 'auto',

  // 在 HTML、Vue、JSX 中每个属性独占一行
  singleAttributePerLine: false,
};
