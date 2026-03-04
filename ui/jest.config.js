export default {
  // 测试环境
  testEnvironment: 'jsdom',
  
  // 测试文件匹配模式
  testMatch: [
    '**/__tests__/**/*.js?(x)',
    '**/__tests__/**/*.jsx?(x)',
    '**/?(*.)+(spec|test).js?(x)',
    '**/?(*.)+(spec|test).jsx?(x)'
  ],
  
  // 模块文件扩展名
  moduleFileExtensions: ['js', 'jsx', 'json'],
  
  // 模块名称映射
  moduleNameMapper: {
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '\\.(gif|ttf|eot|svg|png)$': '<rootDir>/__tests__/__mocks__/fileMock.js'
  },
  
  // 设置文件
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.js'],
  
  // 覆盖率报告
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/**/*.d.ts',
    '!src/main.jsx',
    '!src/index.css'
  ],
  
  // 覆盖率阈值
  coverageThreshold: {
    global: {
      branches: 50,
      functions: 50,
      lines: 50,
      statements: 50
    }
  },
  
  // 转换
  transform: {
    '^.+\\.(js|jsx)$': 'babel-jest'
  },
  
  // 忽略的路径
  testPathIgnorePatterns: [
    '<rootDir>/node_modules/',
    '<rootDir>/dist/',
    '<rootDir>/build/'
  ],
  
  // 详细输出
  verbose: true,
  
  // 显示测试执行时间
  testTimeout: 10000,
  
  // 测试失败后是否退出
  bail: false,
  
  // 是否收集覆盖率
  collectCoverage: true,
  
  // 覆盖率报告目录
  coverageDirectory: '<rootDir>/coverage',
  
  // 覆盖率报告格式
  coverageReporters: ['text', 'lcov', 'html']
};
