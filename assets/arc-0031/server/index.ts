import bodyParser from 'body-parser'
import dotenv from 'dotenv'
import express, { Express } from 'express'
import helmet from 'helmet'
import logger from 'pino-http'
import apiV1Route from './routes/api/v1'

dotenv.config()

const app: Express = express()
const port = process.env.PORT || 3000

app.use(helmet())

app.use(bodyParser.json())
app.use(bodyParser.urlencoded({ extended: true }))

app.use(logger({ prettifier: false }))

app.use(express.static('public'))

app.use('/api/v1', apiV1Route)

app.listen(port, () => {
  console.log(`⚡️[server]: Server is running at http://localhost:${port}`)
})
