import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type ImageSource = {
    type: 'svg' | 'png';
    source: React.ComponentType<React.ComponentProps<'svg'>> | string;
};

type FeatureItem = {
    title: string;
    image: ImageSource;
    description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
    {
        title: 'Plug-and-Play RAG Pipeline',
        image: {
            type: 'png',
            source: require('@site/static/img/rag.png').default,
        },
        description: (
            <>
                Instantly connect your data sources, index content, and leverage LLMs.
                Build powerful AI applications in minutes, not months.
            </>
        ),
    },
    {
        title: 'Multi-Source Knowledge Integration',
        image: {
            type: 'png',
            source: require('@site/static/img/multi_knowledge_source.png').default,
        },
        description: (
            <>
                Seamlessly integrate data from Slack, Google Drive, Confluence, and more.
                Create a unified knowledge base that powers intelligent responses.
            </>
        ),
    },
    {
        title: 'Provider-Agnostic Architecture',
        image: {
            type: 'png',
            source: require('@site/static/img/undraw_docusaurus_tree.svg'),
        },
        description: (
            <>
                Choose from multiple embedding and LLM providers. Avoid vendor lock-in
                and always use the best AI services for your specific needs.
            </>
        ),
    },
    {
        title: 'Advanced Observability',
        image: {
            type: 'svg',
            source: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        },
        description: (
            <>
                Gain deep insights into your AI operations with built-in tracing and metrics.
                Optimize performance and troubleshoot issues with ease.
            </>
        ),
    },
    {
        title: 'Flexible Deployment Options',
        image: {
            type: 'png',
            source: require('@site/static/img/flexible_deployment.png').default,
        },
        description: (
            <>
                Deploy on your local machine, in the cloud, or on Kubernetes.
                Scale effortlessly from development to enterprise-level production.
            </>
        ),
    },
    {
        title: 'Open-Source Community Power',
        image: {
            type: 'svg',
            source: require('@site/static/img/undraw_docusaurus_tree.svg').default,
        },
        description: (
            <>
                Benefit from continuous improvements driven by a vibrant community.
                Extend functionality with plugins and contribute to shaping the future of AI.
            </>
        ),
    },
];

function FeatureImage({ image }: { image: ImageSource }) {
    if (image.type === 'svg') {
        const SvgComponent = image.source as React.ComponentType<React.ComponentProps<'svg'>>;
        return <SvgComponent className={styles.featureSvg} role="img" />;
    } else {
        return (
            <img
                src={image.source as string}
                className={styles.featureSvg}
                alt=""
                role="img"
            />
        );
    }
}

function Feature({ title, image, description }: FeatureItem) {
    return (
        <div className={clsx('col col--4')}>
            <div className="text--center">
                <FeatureImage image={image} />
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