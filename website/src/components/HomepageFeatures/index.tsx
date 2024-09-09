import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
    title: string;
    Svg: React.ComponentType<React.ComponentProps<'svg'>>;
    description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
    {
        title: 'Plug-and-Play RAG Pipeline',
        Svg: require('@site/static/img/rag-pipeline-svg.svg').default,
        description: (
            <>
                Instantly connect your data sources, index content, and leverage LLMs.
                Build powerful AI applications in minutes, not months.
            </>
        ),
    },
    {
        title: 'Multi-Source Knowledge Integration',
        Svg: require('@site/static/img/mutli-source-knowledge.svg').default,
        description: (
            <>
                Seamlessly integrate data from Slack, Google Drive, Confluence, and more.
                Create a unified knowledge base that powers intelligent responses.
            </>
        ),
    },
    {
        title: 'Provider-Agnostic Architecture',
        Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        description: (
            <>
                Choose from multiple embedding and LLM providers. Avoid vendor lock-in
                and always use the best AI services for your specific needs.
            </>
        ),
    },
    {
        title: 'Advanced Observability',
        Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        description: (
            <>
                Gain deep insights into your AI operations with built-in tracing and metrics.
                Optimize performance and troubleshoot issues with ease.
            </>
        ),
    },
    {
        title: 'Flexible Deployment Options',
        Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        description: (
            <>
                Deploy on your local machine, in the cloud, or on Kubernetes.
                Scale effortlessly from development to enterprise-level production.
            </>
        ),
    },
    {
        title: 'Open-Source Community Power',
        Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        description: (
            <>
                Benefit from continuous improvements driven by a vibrant community.
                Extend functionality with plugins and contribute to shaping the future of AI.
            </>
        ),
    },
];

function Feature({title, Svg, description}: FeatureItem) {
    return (
        <div className={clsx('col col--4')}>
            <div className="text--center">
                <Svg className={styles.featureSvg} role="img" />
            </div>
            <div className="text--center padding-horiz--md">
                <Heading as="h3">{title}</Heading>
                <p>{description}</p>
            </div>
        </div>
    );
}

export default function HomepageFeatures(): JSX.Element {
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