module.exports = {
  apps: [
    {
      name: 'admin-portal',
      script: 'pnpm',
      args: 'dev',
      cwd: './admin-portal',
      env: {
        NODE_ENV: 'development',
        PORT: 5173,
      },
      env_production: {
        NODE_ENV: 'production',
        PORT: 5173,
      },
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '1G',
      error_file: './logs/admin-portal-error.log',
      out_file: './logs/admin-portal-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss',
    },
    {
      name: 'merchant-portal',
      script: 'pnpm',
      args: 'dev',
      cwd: './merchant-portal',
      env: {
        NODE_ENV: 'development',
        PORT: 5174,
      },
      env_production: {
        NODE_ENV: 'production',
        PORT: 5174,
      },
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '1G',
      error_file: './logs/merchant-portal-error.log',
      out_file: './logs/merchant-portal-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss',
    },
    {
      name: 'website',
      script: 'pnpm',
      args: 'dev',
      cwd: './website',
      env: {
        NODE_ENV: 'development',
        PORT: 5175,
      },
      env_production: {
        NODE_ENV: 'production',
        PORT: 5175,
      },
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '1G',
      error_file: './logs/website-error.log',
      out_file: './logs/website-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss',
    },
  ],
}


