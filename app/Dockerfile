FROM node:18.0.0-alpine

WORKDIR /usr/src/app

COPY . .

RUN npm ci --omit=dev

EXPOSE 3000

CMD npm run start