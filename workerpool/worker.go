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

type Job interface {
    Do()
}


type Worker struct {
    jobsChan    chan Job        /* Chan to receive new jobs */
    pool        chan chan Job   /* */
    stopChan    chan bool       /* Stop the worker*/
}


func NewWorker(pool chan chan Job) (*Worker) {
    w := &Worker {
        jobsChan:   make(chan Job),
        pool:       pool,
        stopChan:   make(chan bool),
    }
    return w
}

func (w * Worker) Start() {
    go func(){
        for {
            w.pool <- w.jobsChan
            select {
            case job, ok := <-w.jobsChan:
                if !ok {
                    w.stopChan <- true
                    return
                } 
                if job != nil {
                    job.Do()
                }
            }
        }
    }()
}

func (w * Worker) Stop() {
    if w.Stopped() {
        return
    }
    close(w.jobsChan)
    <-w.stopChan
}

func (w * Worker) Stopped (bool) {
    select {
    case <- w.stopChan:
        return true
    default:
    }
    return false
}

