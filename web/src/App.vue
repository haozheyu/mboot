<template>
  <!-- =========================================================
       Loading
       ========================================================= -->
  <div
      v-if="authMode === 'loading'"
      class="relative flex min-h-screen items-center justify-center overflow-hidden bg-slate-950 px-4"
  >
    <div class="pointer-events-none absolute inset-0">
      <div
          class="absolute left-1/2 top-[-280px] h-[620px] w-[620px]
               -translate-x-1/2 rounded-full bg-blue-600/10 blur-3xl"
      ></div>

      <div
          class="absolute bottom-[-300px] right-[-200px] h-[520px] w-[520px]
               rounded-full bg-cyan-500/10 blur-3xl"
      ></div>

      <div class="grid-background absolute inset-0"></div>
    </div>

    <div class="relative flex flex-col items-center">
      <div
          class="mb-5 flex h-14 w-14 items-center justify-center
               rounded-2xl bg-blue-600 text-lg font-bold text-white
               shadow-xl shadow-blue-600/20"
      >
        PX
      </div>

      <div class="flex items-center gap-3 text-sm text-slate-400">
        <span
            class="h-4 w-4 animate-spin rounded-full
                 border-2 border-slate-600 border-t-blue-500"
        ></span>

        正在连接管理后台...
      </div>
    </div>
  </div>

  <!-- =========================================================
       Login / Setup
       ========================================================= -->
  <div
      v-else-if="authMode !== 'ready'"
      class="relative flex min-h-screen items-center justify-center
           overflow-hidden bg-slate-950 px-4 py-10"
  >
    <!-- Background -->
    <div class="pointer-events-none absolute inset-0">
      <div
          class="absolute left-1/2 top-[-320px] h-[680px] w-[680px]
               -translate-x-1/2 rounded-full bg-blue-600/10 blur-3xl"
      ></div>

      <div
          class="absolute bottom-[-300px] right-[-220px] h-[560px] w-[560px]
               rounded-full bg-cyan-500/10 blur-3xl"
      ></div>

      <div
          class="absolute bottom-[-260px] left-[-240px] h-[480px] w-[480px]
               rounded-full bg-indigo-500/[0.08] blur-3xl"
      ></div>

      <div class="grid-background absolute inset-0"></div>
    </div>

    <!-- Login Container -->
    <div class="relative w-full max-w-[420px]">
      <!-- Logo -->
      <div class="mb-8 text-center">
        <div
            class="mx-auto mb-4 flex h-14 w-14 items-center justify-center
                 rounded-2xl bg-blue-600 text-xl font-bold text-white
                 shadow-xl shadow-blue-600/25"
        >
          PX
        </div>

        <h1 class="text-2xl font-semibold tracking-tight text-white">
          &nbsp;
        </h1>

        <p class="mt-7 text-sm text-slate-400">
          Network Boot Management Platform
        </p>
      </div>

      <!-- Login Card -->
      <div
          class="rounded-2xl border border-white/10 bg-white/[0.06]
               p-7 shadow-2xl shadow-black/20 backdrop-blur-xl"
      >
        <div class="mb-6">
          <h2 class="text-lg font-semibold text-white">
            {{
              authMode === 'setup'
                  ? '初始化管理账号'
                  : '登录管理控制台'
            }}
          </h2>

          <p class="mt-1.5 text-sm leading-6 text-slate-400">
            {{
              authMode === 'setup'
                  ? '首次使用 PXE Console，请创建系统管理员账号。'
                  : '请输入管理员账号和密码继续访问。'
            }}
          </p>
        </div>

        <div class="space-y-5">
          <!-- Username -->
          <div>
            <label
                for="username"
                class="mb-2 block text-sm font-medium text-slate-300"
            >
              用户名
            </label>

            <input
                id="username"
                v-model.trim="username"
                class="auth-input"
                placeholder="请输入用户名"
                autocomplete="username"
                spellcheck="false"
            />

            <p
                v-if="authMode === 'setup'"
                class="mt-2 text-xs leading-5 text-slate-500"
            >
              3-32 位，仅支持字母、数字、点、下划线、短横线和 @。
            </p>
          </div>

          <!-- Password -->
          <div>
            <label
                for="password"
                class="mb-2 block text-sm font-medium text-slate-300"
            >
              密码
            </label>

            <div class="relative">
              <input
                  id="password"
                  v-model="password"
                  class="auth-input pr-11"
                  :type="showPassword ? 'text' : 'password'"
                  :placeholder="
                  authMode === 'setup'
                    ? '请输入至少 8 位密码'
                    : '请输入密码'
                "
                  :autocomplete="
                  authMode === 'setup'
                    ? 'new-password'
                    : 'current-password'
                "
                  @keydown.enter="submitAuth"
              />

              <button
                  type="button"
                  class="absolute inset-y-0 right-0 flex w-11
                       items-center justify-center text-slate-500
                       transition hover:text-slate-300"
                  :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                  @click="showPassword = !showPassword"
              >
                <EyeOff
                    v-if="showPassword"
                    class="h-[17px] w-[17px]"
                />

                <Eye
                    v-else
                    class="h-[17px] w-[17px]"
                />
              </button>
            </div>
          </div>

          <!-- Login Button -->
          <button
              type="button"
              class="flex h-11 w-full items-center justify-center
                   rounded-lg bg-blue-600 px-4 text-sm font-medium
                   text-white shadow-lg shadow-blue-600/10
                   transition
                   hover:bg-blue-500
                   focus:outline-none
                   focus:ring-4 focus:ring-blue-500/20
                   active:bg-blue-700
                   disabled:cursor-not-allowed
                   disabled:opacity-50"
              :disabled="authBusy"
              @click="submitAuth"
          >
            <LoaderCircle
                v-if="authBusy"
                class="mr-2 h-4 w-4 animate-spin"
            />

            {{
              authBusy
                  ? '处理中...'
                  : authMode === 'setup'
                      ? '创建管理员'
                      : '登录控制台'
            }}
          </button>

          <!-- Error -->
          <div
              v-if="authError"
              class="flex gap-2.5 rounded-lg border border-red-500/20
                   bg-red-500/10 px-3.5 py-3 text-sm
                   leading-5 text-red-300"
          >
            <CircleAlert
                class="mt-0.5 h-4 w-4 shrink-0"
            />

            <span class="min-w-0 break-words">
              {{ authError }}
            </span>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div
          class="mt-66 flex items-center justify-center gap-2
               text-xs text-slate-600"
      >
        <Server class="h-3.5 w-3.5" />

        PXE Network Boot Management
      </div>
    </div>
  </div>

  <!-- =========================================================
       Main Application
       ========================================================= -->
  <div
      v-else
      class="min-h-screen bg-slate-50"
  >
    <!-- =======================================================
         Desktop Sidebar
         ======================================================= -->
    <aside
        class="fixed inset-y-0 left-0 z-30 hidden w-64
             flex-col border-r border-slate-800
             bg-slate-950 lg:flex"
    >
      <!-- Logo -->
      <div
          class="flex h-16 shrink-0 items-center
               border-b border-slate-800 px-5"
      >
        <div class="flex items-center gap-3">
          <div
              class="flex h-9 w-9 shrink-0 items-center
                   justify-center rounded-xl bg-blue-600
                   text-xs font-bold text-white
                   shadow-lg shadow-blue-600/20"
          >
            PX
          </div>

          <div class="min-w-0">
            <div class="truncate text-sm font-semibold text-white">
              PXE Console
            </div>

            <div class="truncate text-xs text-slate-500">
              Network Boot Manager
            </div>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <div
          class="flex-1 overflow-y-auto px-3 py-4
               sidebar-scrollbar"
      >
        <div
            class="mb-2 px-3 text-[10px] font-semibold
                 uppercase tracking-[0.16em] text-slate-600"
        >
          Navigation
        </div>

        <nav class="space-y-1">
          <RouterLink
              v-for="item in nav"
              :key="item.path"
              :to="item.path"
              class="group flex items-center gap-3 rounded-lg
                   px-3 py-2.5 text-sm font-medium
                   text-slate-400 transition-all duration-150
                   hover:bg-slate-900 hover:text-white"
              active-class="nav-link-active"
          >
            <component
                :is="item.icon"
                class="h-[18px] w-[18px] shrink-0"
            />

            <span class="truncate">
              {{ item.name }}
            </span>
          </RouterLink>
        </nav>
      </div>

      <!-- Sidebar Footer -->
      <div
          class="shrink-0 border-t border-slate-800 p-3"
      >
        <a
            class="flex items-center gap-3 rounded-lg
                 px-3 py-2.5 text-sm text-slate-500
                 transition hover:bg-slate-900
                 hover:text-slate-300"
            href="https://github.com/haozheyu/mboot"
            target="_blank"
            rel="noreferrer"
        >
          <Github class="h-[18px] w-[18px]" />

          <span>GitHub</span>

          <ExternalLink
              class="ml-auto h-3.5 w-3.5"
          />
        </a>

        <div
            class="mt-2 flex items-center gap-2 px-3
                 py-1 text-[11px] text-slate-700"
        >
          <span
              class="h-1.5 w-1.5 rounded-full bg-emerald-500"
          ></span>

          Management Console
        </div>
      </div>
    </aside>

    <!-- =======================================================
         Mobile Sidebar
         ======================================================= -->
    <Transition name="overlay">
      <div
          v-if="mobileOpen"
          class="fixed inset-0 z-50 lg:hidden"
      >
        <!-- Overlay -->
        <button
            class="absolute inset-0 bg-slate-950/50 backdrop-blur-[2px]"
            aria-label="关闭导航"
            @click="mobileOpen = false"
        ></button>

        <!-- Sidebar -->
        <aside
            class="absolute inset-y-0 left-0 flex w-72
                 max-w-[84vw] flex-col border-r
                 border-slate-800 bg-slate-950
                 shadow-2xl"
        >
          <!-- Logo -->
          <div
              class="flex h-16 shrink-0 items-center
                   justify-between border-b
                   border-slate-800 px-4"
          >
            <div class="flex items-center gap-3">
              <div
                  class="flex h-9 w-9 items-center
                       justify-center rounded-xl bg-blue-600
                       text-xs font-bold text-white"
              >
                PX
              </div>

              <div>
                <div class="text-sm font-semibold text-white">
                  PXE Console
                </div>

                <div class="text-xs text-slate-500">
                  Network Boot Manager
                </div>
              </div>
            </div>

            <button
                type="button"
                class="flex h-9 w-9 items-center justify-center
                     rounded-lg border border-slate-800
                     text-slate-400 transition
                     hover:bg-slate-900 hover:text-white"
                aria-label="关闭导航"
                @click="mobileOpen = false"
            >
              <X class="h-4 w-4" />
            </button>
          </div>

          <!-- Mobile Navigation -->
          <div
              class="flex-1 overflow-y-auto px-3 py-4"
          >
            <div
                class="mb-2 px-3 text-[10px] font-semibold
                     uppercase tracking-[0.16em]
                     text-slate-600"
            >
              Navigation
            </div>

            <nav class="space-y-1">
              <RouterLink
                  v-for="item in nav"
                  :key="item.path"
                  :to="item.path"
                  class="flex items-center gap-3 rounded-lg
                       px-3 py-2.5 text-sm font-medium
                       text-slate-400 transition
                       hover:bg-slate-900 hover:text-white"
                  active-class="nav-link-active"
                  @click="mobileOpen = false"
              >
                <component
                    :is="item.icon"
                    class="h-[18px] w-[18px]"
                />

                {{ item.name }}
              </RouterLink>
            </nav>
          </div>

          <!-- Mobile Footer -->
          <div
              class="shrink-0 border-t
                   border-slate-800 p-3"
          >
            <a
                class="flex items-center gap-3 rounded-lg
                     px-3 py-2.5 text-sm text-slate-500
                     transition hover:bg-slate-900
                     hover:text-slate-300"
                href="https://github.com/sky22333/netboot"
                target="_blank"
                rel="noreferrer"
            >
              <Github class="h-[18px] w-[18px]" />

              GitHub

              <ExternalLink
                  class="ml-auto h-3.5 w-3.5"
              />
            </a>
          </div>
        </aside>
      </div>
    </Transition>

    <!-- =======================================================
         Content
         ======================================================= -->
    <div class="min-h-screen lg:pl-64">
      <!-- Header -->
      <header
          class="sticky top-0 z-20 flex h-16
               items-center justify-between
               border-b border-slate-200
               bg-white/90 px-4
               backdrop-blur-xl
               sm:px-5 lg:px-6"
      >
        <!-- Header Left -->
        <div class="flex min-w-0 items-center gap-3">
          <!-- Mobile Menu -->
          <button
              type="button"
              class="flex h-9 w-9 shrink-0
                   items-center justify-center
                   rounded-lg border border-slate-200
                   bg-white text-slate-600
                   shadow-sm transition
                   hover:bg-slate-50
                   hover:text-slate-950
                   lg:hidden"
              aria-label="打开导航"
              @click="mobileOpen = true"
          >
            <Menu class="h-4 w-4" />
          </button>

          <!-- Page Title -->
          <div class="min-w-0">
            <div
                class="text-[11px] font-medium
                     uppercase tracking-wider
                     text-slate-400"
            >
              Current Page
            </div>

            <div
                class="truncate text-base font-semibold
                     leading-5 text-slate-900"
            >
              {{ title }}
            </div>
          </div>
        </div>

        <!-- Header Right -->
        <div class="flex items-center gap-2">
          <div
              class="hidden items-center gap-2
                   rounded-full border border-emerald-200
                   bg-emerald-50 px-3 py-1.5
                   text-xs font-medium text-emerald-700
                   sm:flex"
          >
            <span
                class="h-1.5 w-1.5 rounded-full
                     bg-emerald-500"
            ></span>

            服务已连接
          </div>

          <button
              type="button"
              class="inline-flex h-9 items-center
                   justify-center gap-2 rounded-lg
                   border border-slate-200 bg-white
                   px-3.5 text-sm font-medium
                   text-slate-700 shadow-sm
                   transition hover:border-slate-300
                   hover:bg-slate-50
                   hover:text-slate-950
                   focus:outline-none
                   focus:ring-4 focus:ring-slate-100
                   disabled:cursor-not-allowed
                   disabled:opacity-60"
              :disabled="refreshing"
              @click="refresh"
          >
            <RefreshCw
                class="h-4 w-4"
                :class="{ 'animate-spin': refreshing }"
            />

            <span class="hidden sm:inline">
              {{ refreshing ? '正在刷新' : '刷新状态' }}
            </span>
          </button>
        </div>
      </header>

      <!-- Main -->
      <main
          class="mx-auto w-full max-w-[1600px]
               p-4 sm:p-6 lg:p-8"
      >
        <RouterView v-slot="{ Component }">
          <Transition
              name="page"
              mode="out-in"
          >
            <component :is="Component" />
          </Transition>
        </RouterView>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  onMounted,
  ref
} from 'vue'

import {
  useRoute
} from 'vue-router'

import {
  Activity,
  CircleAlert,
  ExternalLink,
  Eye,
  EyeOff,
  Files,
  Gauge,
  Github,
  HardDrive,
  ListTree,
  LoaderCircle,
  Menu,
  Network,
  RefreshCw,
  ScrollText,
  Server,
  Settings,
  TerminalSquare,
  Users,
  X
} from 'lucide-vue-next'

import { api } from './lib/api'

/**
 * ============================================================
 * Navigation
 * ============================================================
 */
const nav = [
  {
    path: '/',
    name: '仪表盘',
    icon: Gauge
  },
  {
    path: '/config',
    name: '服务配置',
    icon: Settings
  },
  {
    path: '/clients',
    name: '客户端',
    icon: Network
  },
  {
    path: '/menus',
    name: '启动菜单',
    icon: ListTree
  },
  {
    path: '/files',
    name: '文件管理',
    icon: Files
  },
  {
    path: '/netboot',
    name: 'netboot.xyz',
    icon: HardDrive
  },
  {
    path: '/actions',
    name: '操作菜单',
    icon: TerminalSquare
  },
  {
    path: '/users',
    name: '用户',
    icon: Users
  },
  {
    path: '/logs',
    name: '日志',
    icon: ScrollText
  },
  {
    path: '/diagnostics',
    name: '系统诊断',
    icon: Activity
  }
]

/**
 * ============================================================
 * Router
 * ============================================================
 */
const route = useRoute()

const title = computed(() => {
  return (
      nav.find((item) => item.path === route.path)?.name ??
      'PXE Console'
  )
})

/**
 * ============================================================
 * Authentication State
 * ============================================================
 */
const authMode = ref<
    'loading' |
    'setup' |
    'login' |
    'ready'
    >('loading')

const username = ref('admin')
const password = ref('')

const authError = ref('')
const authBusy = ref(false)

const showPassword = ref(false)

/**
 * ============================================================
 * UI State
 * ============================================================
 */
const refreshing = ref(false)
const mobileOpen = ref(false)

/**
 * ============================================================
 * Error normalization
 *
 * 防止后端 / Nginx / Vite 返回 HTML 时，
 * 把整段 <!doctype html> 显示到登录页面。
 * ============================================================
 */
function normalizeError(
    error: unknown,
    fallback = '操作失败'
): string {
  if (!(error instanceof Error)) {
    return fallback
  }

  const message = error.message?.trim()

  if (!message) {
    return fallback
  }

  const lower = message.toLowerCase()

  if (
      lower.includes('<!doctype html') ||
      lower.includes('<html') ||
      lower.includes('<head') ||
      lower.includes('/@vite/client')
  ) {
    return '管理后台接口返回异常，请检查 API 地址或反向代理配置'
  }

  /**
   * 避免异常信息过长破坏 UI。
   */
  if (message.length > 300) {
    return `${message.slice(0, 300)}...`
  }

  return message
}

/**
 * ============================================================
 * Refresh
 * ============================================================
 */
function refresh() {
  if (refreshing.value) {
    return
  }

  refreshing.value = true

  window.dispatchEvent(
      new CustomEvent('pxe-refresh')
  )

  window.setTimeout(() => {
    refreshing.value = false
  }, 900)
}

/**
 * ============================================================
 * Check Authentication
 * ============================================================
 */
async function checkAuth() {
  authError.value = ''

  try {
    /**
     * 1. 检查系统是否已经存在管理员
     */
    const setup = await api<{
      has_user: boolean
    }>('/setup/status')

    /**
     * 2. 未初始化
     */
    if (!setup.has_user) {
      authMode.value = 'setup'
      return
    }

    /**
     * 3. 已初始化，检查当前 Session
     */
    await api('/status')

    authMode.value = 'ready'
  } catch (error) {
    /**
     * 当前后端 API 的行为是：
     * 未登录时 /status 抛出错误。
     *
     * 所以只要管理员已经存在，但 /status
     * 失败，就进入登录页面。
     */
    authError.value = normalizeError(
        error,
        '无法连接到后端服务'
    )

    authMode.value = 'login'
  }
}

/**
 * ============================================================
 * Login / Setup
 * ============================================================
 */
async function submitAuth() {
  if (authBusy.value) {
    return
  }

  /**
   * 基础前端校验
   */
  if (!username.value.trim()) {
    authError.value = '请输入用户名'
    return
  }

  if (!password.value) {
    authError.value = '请输入密码'
    return
  }

  if (
      authMode.value === 'setup' &&
      password.value.length < 8
  ) {
    authError.value = '管理员密码不能少于 8 位'
    return
  }

  authBusy.value = true
  authError.value = ''

  try {
    /**
     * 首次初始化
     */
    if (authMode.value === 'setup') {
      await api('/setup', {
        method: 'POST',
        body: JSON.stringify({
          username: username.value,
          password: password.value
        })
      })
    }

    /**
     * 登录
     */
    await api('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        username: username.value,
        password: password.value
      })
    })

    /**
     * 清除密码
     */
    password.value = ''

    /**
     * 进入后台
     */
    authMode.value = 'ready'
  } catch (error) {
    authError.value = normalizeError(
        error,
        '登录失败'
    )
  } finally {
    authBusy.value = false
  }
}

/**
 * ============================================================
 * Mounted
 * ============================================================
 */
onMounted(() => {
  checkAuth()
})
</script>

<style scoped>
/*
 * ============================================================
 * Login Background
 * ============================================================
 */

.grid-background {
  opacity: 0.035;
  background-image:
      linear-gradient(
          to right,
          #ffffff 1px,
          transparent 1px
      ),
      linear-gradient(
          to bottom,
          #ffffff 1px,
          transparent 1px
      );

  background-size: 40px 40px;

  mask-image:
      linear-gradient(
          to bottom,
          black,
          transparent 90%
      );
}

/*
 * ============================================================
 * Authentication Input
 * ============================================================
 */

.auth-input {
  width: 100%;
  height: 44px;

  padding-left: 14px;
  padding-right: 14px;

  border-radius: 8px;

  border:
      1px solid
      rgba(255, 255, 255, 0.1);

  background:
      rgba(255, 255, 255, 0.06);

  color: #ffffff;

  font-size: 14px;

  outline: none;

  transition:
      border-color 150ms ease,
      box-shadow 150ms ease,
      background-color 150ms ease;
}

.auth-input::placeholder {
  color: #64748b;
}

.auth-input:hover {
  border-color:
      rgba(255, 255, 255, 0.2);
}

.auth-input:focus {
  border-color: #3b82f6;

  box-shadow:
      0 0 0 4px
      rgba(59, 130, 246, 0.1);

  background:
      rgba(255, 255, 255, 0.08);
}

/*
 * 浏览器自动填充密码/用户名时，
 * 避免输入框突然变成白色。
 */
.auth-input:-webkit-autofill,
.auth-input:-webkit-autofill:hover,
.auth-input:-webkit-autofill:focus {
  -webkit-text-fill-color: #ffffff;

  -webkit-box-shadow:
      0 0 0 1000px
      #172033 inset;

  transition:
      background-color
      9999s ease-in-out 0s;
}

/*
 * ============================================================
 * Active Navigation
 * ============================================================
 */

.nav-link-active {
  color: #60a5fa !important;

  background:
      rgba(37, 99, 235, 0.14) !important;

  box-shadow:
      inset 3px 0 0 #3b82f6;
}

/*
 * ============================================================
 * Sidebar Scrollbar
 * ============================================================
 */

.sidebar-scrollbar {
  scrollbar-width: thin;
  scrollbar-color:
      #334155
      transparent;
}

.sidebar-scrollbar::-webkit-scrollbar {
  width: 5px;
}

.sidebar-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-scrollbar::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: #334155;
}

.sidebar-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/*
 * ============================================================
 * Page Transition
 * ============================================================
 */

.page-enter-active,
.page-leave-active {
  transition:
      opacity 150ms ease,
      transform 150ms ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}

/*
 * ============================================================
 * Mobile Overlay
 * ============================================================
 */

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 180ms ease;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

/*
 * ============================================================
 * Accessibility
 * ============================================================
 */

@media (prefers-reduced-motion: reduce) {
  .page-enter-active,
  .page-leave-active,
  .overlay-enter-active,
  .overlay-leave-active {
    transition: none;
  }
}
</style>