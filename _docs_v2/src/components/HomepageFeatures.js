import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'No reflection',
    description: (
        <ul>
            <li>The json encoding is code-generated, optimized and uses <a href="https://github.com/ogen-go/jx/">jx</a> for speed and overcoming encoding/json limitations</li>
            <li>Validation is code-generated according to specification</li>
        </ul>
    ),
  },
  {
    title: 'No boilerplate',
    description: (
        <ul>
            <li>Structures are generated from OpenAPI v3 specification</li>
            <li>Arguments, headers, url queries are parsed according to specification into structures</li>
            <li>String formats like <code>uuid</code>, <code>date</code>, <code>date-time</code>, <code>uri</code> are represented by go types directly</li>
            <li>Sum types are generated for OneOf, with discriminator or implicit type inference</li>
            <li>Optional and nullable are supported <b>without pointers</b> if possible</li>
        </ul>
    ),
  },
  {
    title: 'OpenTelemetry',
    description: (
      <>
        Tracing and metrics support that is compatible with OpenTelemtry
      </>
    ),
  },
];

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
