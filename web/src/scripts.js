'use strict'

import { initRefreshButton } from './fn'
import TableData from './TableData'
import Fetcher from './Fetcher'
import Error from './Error'
import state from './state'
import cfg from './config'

const td = new TableData(new Fetcher(), new Error())

const timeout = +state.urlParams.get(cfg.queryTimeout) || 0
initRefreshButton(timeout)

td.fetchStats()
  .finally(td.show.bind(td))
  .finally(() => {
    if (timeout) {
      const repeatRequest = () => setTimeout(() => td.fetchStats().finally(repeatRequest), timeout)
      repeatRequest()
    }
  })
