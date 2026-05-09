/** @type {import('next').NextConfig} */
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '/gptimg'

const nextConfig = {
  output: 'export',
  basePath,
  images: {
    unoptimized: true,
  },
}

module.exports = nextConfig
