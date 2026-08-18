/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/antd/**/*.{js,jsx,ts,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f7fbf0',    // 最浅的绿色背景
          100: '#eef7e0',   // 浅绿色背景
          200: '#ddeec0',   // 浅绿色边框
          300: '#c2e090',   // 中等绿色
          400: '#a5d160',   // 中深绿色
          500: '#8CC020',   // 主色调
          600: '#7aad1c',   // 深绿色（悬停）
          700: '#689a18',   // 更深的绿色
          800: '#567714',   // 深绿色（激活）
          900: '#445410',   // 最深的绿色
        },
        accent: {
          light: '#f8fafc',  // 浅灰背景
          medium: '#e2e8f0', // 中灰边框
          dark: '#64748b',   // 深灰文字
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        'primary': '0 4px 14px 0 rgba(140, 192, 32, 0.15)',
        'primary-lg': '0 10px 25px 0 rgba(140, 192, 32, 0.2)',
      }
    },
  },
  plugins: [],
  corePlugins: {
    preflight: false,
  }
}
