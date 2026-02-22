import dotenv from 'dotenv'
import { type Express } from 'express'
import express from 'express'
import helmet from 'helmet'
import morgan from 'morgan'

import { env } from './env'
import apiRoutes from './routes/api/v1'

dotenv.config()

const app: Express = express()
const port = env.PORT || 3000

app.use(helmet())
app.use(express.json())
app.use(express.urlencoded({ extended: true }))
app.use(express.static('public'))
app.use(morgan('[:date[iso]] ":method :url HTTP/:http-version" :status :res[content-length] ":referrer" ":user-agent"'))

app.use('/api/v1', apiRoutes)

app.listen(port, () => {
  console.log(`⚡️[server]: Server is running at http://localhost:${port}`)
})
