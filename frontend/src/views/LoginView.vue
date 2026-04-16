<template>
  <div class="login-container">
    <div class="login-box">
      <h1 class="login-title">业余无线电模拟考试系统</h1>
      
      <div class="input-group">
        <el-input
          v-model="username"
          placeholder="请输入用户名"
          size="large"
          prefix-icon="User"
          clearable
        />
      </div>
      
      <div class="input-group">
        <el-input
          v-model="idCard"
          placeholder="请输入身份证号"
          size="large"
          prefix-icon="Ticket"
          clearable
        />
      </div>
      
      <div class="button-group">
        <el-button type="primary" size="large" @click="handleLogin" :loading="loading">
          登录
        </el-button>
      </div>
      
      <div class="copyright">
        {{ copyright }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { UserService } from '@/api'

const router = useRouter()
const userStore = useUserStore()

const username = ref('')
const idCard = ref('')
const loading = ref(false)
const copyright = 'Copyright © 2024-2025 BA4RHH. All rights reserved.'

const handleLogin = async () => {
  if (!username.value.trim()) {
    ElMessage.warning('请输入用户名')
    return
  }
  
  if (!idCard.value.trim()) {
    ElMessage.warning('请输入身份证号')
    return
  }

  loading.value = true
  
  try {
    const result = await UserService.Login(username.value, idCard.value)
    
    if (result && result.user_info) {
      userStore.setUser({
        user_id: result.user_info.id,
        username: result.user_info.username,
        password: '',
        last_login: result.user_info.last_login as any
      })
      ElMessage.success('登录成功')
      router.push('/home')
    } else {
      ElMessage.error('用户名或身份证号错误')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '登录失败，请检查网络连接')
    console.error('Login error:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 450px;
  padding: 40px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.login-title {
  text-align: center;
  color: #333;
  margin-bottom: 40px;
  font-size: 24px;
  font-weight: 600;
}

.input-group {
  margin-bottom: 20px;
}

.button-group {
  margin-top: 30px;
  text-align: center;
}

.button-group .el-button {
  width: 100%;
}

.copyright {
  text-align: center;
  margin-top: 30px;
  color: #999;
  font-size: 12px;
}
</style>
