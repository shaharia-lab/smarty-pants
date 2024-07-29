# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm ci --ignore-scripts

# Copy all files
COPY . .

# Build the Next.js application
RUN npm run build

# Production stage
FROM node:20-alpine AS runner

WORKDIR /app

# Set environment to production
ENV NODE_ENV production

# Copy necessary files from builder stage
COPY --from=builder next.config.mjs ./
COPY --from=builder public ./public
COPY --from=builder .next ./.next
COPY --from=builder node_modules ./node_modules
COPY --from=builder package.json ./package.json

# Run the Next.js application
CMD ["npm", "start"]