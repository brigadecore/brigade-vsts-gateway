/**
 * brigadier is the client-facing library for Brigade.
 *
 * Objects in this library are available to `brigade.js` scripts.
 */

/** */

import * as jobImpl from "./job";
import * as groupImpl from "./group";
import * as eventsImpl from "./events";
import { JobRunner } from "./k8s";

// These are filled by the 'fire' event handler.
let currentEvent = null;
let currentProject = null;

/**
 * events is the main event registry.
 *
 * New event handlers can be registered using `events.on(name: string, (e: BrigadeEvent, p: Project) => {})`.
 * where the `name` is the event name, and the callback is the function to be
 * executed when the event is triggered.
 */
export let events = new eventsImpl.EventRegistry();

/**
 * fire triggers an event.
 *
 * The fire() function takes a BrigadeEvent (the event to be triggered) and a
 * Project (the owner project). If an event handler is found, it is executed.
 * If no event handler is found, nothing happens.
 */
export function fire(e: eventsImpl.BrigadeEvent, p: eventsImpl.Project) {
  currentEvent = e;
  currentProject = p;
  events.fire(e, p);
}

/**
 * Job describes a particular job.
 *
 * A Job always has a name and an image. The name is used to reference this
 * job in relation to other jobs in the same event. The image corresponds to a
 * container image that will be executed as part of this job.
 *
 * A Job may also have one or more tasks associated with it. Tasks are run
 * (in order) inside of the image. When no tasks are supplied, the image is
 * executed as-is.
 */
export class Job extends jobImpl.Job {
  jr: JobRunner;

  run(): Promise<jobImpl.Result> {
    this.jr = new JobRunner(this, currentEvent, currentProject);
    this._podName = this.jr.name;
    return this.jr.run().catch(err => {
      // Wrap the message to give clear context.
      console.error(err);
      let msg = `job ${this.name}(${this.jr.name}): ${err}`;
      return Promise.reject(new Error(msg));
    });
  }

  logs(): Promise<string> {
    return this.jr.logs();
  }
}

/**
 * Group describes a collection of associated jobs.
 *
 * A group of jobs can be executed in two ways:
 *   - In parallel as runAll()
 *   - In serial as runEach()
 */
export class Group extends groupImpl.Group {
  // This seems to be how you expose an existing class as an export.
}

/**
 * ErrorReport describes an error in the runtime handling of a Brigade script.
 */
export class ErrorReport {
  /**
   * cause is the BrigadeEvent that caused the error.
   */
  public cause: eventsImpl.BrigadeEvent;
  /**
   * reason is the message that the error reporter received that describes the error.
   *
   * It may be empty if no error description was provided.
   */
  public reason?: any;
}
