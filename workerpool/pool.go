////////////////////////////////////////////////////////////////////////////////
// 
// (C) 2011 Kevin Druelle <kevin@druelle.info>
//
// this software is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This software is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this software.  If not, see <http://www.gnu.org/licenses/>.
// 
///////////////////////////////////////////////////////////////////////////////

package workerpool


type WorkerPool struct {
    maxWorkers      int
    workers         []*Worker
    jobQueue        chan Job
    pool            chan chan Job
    stopChan        chan bool
}


func NewWorkerPool(workers, queueLength int) (*WorkerPool) {
    p := &WorkerPool{
        maxWorkers: workers,
        jobQueue:   make(chan Job, queueLength),
        pool:       make(chan chan Job, workers),
        stopChan:   make(chan bool),
    }
    return p
}

func (p * WorkerPool) Stop() {
    if p.Stopped() {
        return
    }
    for _, w := range p.workers {
        w.Stop()
    }
    close(p.jobQueue)
    <-p.stopChan
}

func (p * WorkerPool) Stopped() (bool) {
    select {
    case <-p.stopChan:
        return true;
    default:
    }
    return false
}

func (p * WorkerPool) Handle(job Job) {
    p.jobQueue <- job
}

func (p * WorkerPool) Run() {
    for i := 0; i < p.maxWorkers; i++ {
        w := NewWorker(p.pool)
        w.Start()
        p.workers = append(p.workers, w)
    }
    p.dispatch()
}

func (p * WorkerPool) dispatch() {
    go func () {
        for {
            select {
            case job, ok := <-p.jobQueue:
                if !ok {
                    p.stopChan <- true
                    return
                }
                if job != nil {
                    go func (job Job) {
                        jobChan := <-p.pool
                        jobChan <- job
                    }(job)
                }
            }
        }
    }()
}


