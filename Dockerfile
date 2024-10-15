FROM node:12.12-alpine

RUN apk --no-cache add make gcc g++ python git

COPY jest.config.json package-lock.json package.json ./

RUN npm i

COPY .env ./

COPY src/ ./src/

ENV NODE_ENV=production

CMD ["npm", "run", "start"]