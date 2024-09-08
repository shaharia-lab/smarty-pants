import {AvailableDatasource} from '@/types/datasource';

export const availableDatasources: AvailableDatasource[] = [
    {
        id: 'slack',
        name: 'Slack',
        imageUrl: '/images/slack_icon.webp',
        description: 'Connect and sync your Slack workspace',
        configurationUrl: '/datasources/slack/add',
    },
];

export function getDatasourceById(id: string): AvailableDatasource | undefined {
    return availableDatasources.find(datasource => datasource.id === id);
}