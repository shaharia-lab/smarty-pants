'use client';

import React, { useEffect, useState } from 'react';
import { ArcElement, BarElement, CategoryScale, Chart as ChartJS, Legend, LinearScale, Title, Tooltip } from 'chart.js';
import { Pie } from 'react-chartjs-2';
import { Loader2 } from 'lucide-react';
import AuthService from "@/services/authService";
import axios from "axios";
import { ApiError, createApiService, AnalyticsOverview } from "@/services/apiService";

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement, Title);

const Dashboard = () => {
    const [analyticsData, setAnalyticsData] = useState<AnalyticsOverview | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const apiService = createApiService(AuthService);

    useEffect(() => {
        const source = axios.CancelToken.source();

        const fetchAnalytics = async () => {
            try {
                const data = await apiService.getAnalyticsOverview(source.token);
                setAnalyticsData(data);
            } catch (error) {
                if (error instanceof ApiError) {
                    // Handle specific API errors
                } else {
                    console.error('Error fetching analytics data:', error);
                }
            } finally {
                setIsLoading(false);
            }
        };

        fetchAnalytics();

        return () => {
            source.cancel('Component unmounted');
        };
    }, []);

    if (isLoading) {
        return (
            <div className="flex justify-center items-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-gray-800"/>
            </div>
        );
    }

    if (!analyticsData) {
        return <div className="text-center text-red-500">Failed to load analytics data</div>;
    }

    const chartColors = [
        'rgba(75, 192, 192, 0.8)',
        'rgba(255, 99, 132, 0.8)',
        'rgba(255, 206, 86, 0.8)',
        'rgba(54, 162, 235, 0.8)',
        'rgba(153, 102, 255, 0.8)',
        'rgba(255, 159, 64, 0.8)',
    ];

    const datasourcesByTypeData = {
        labels: Object.keys(analyticsData.datasources.total_datasources_by_type),
        datasets: [
            {
                data: Object.values(analyticsData.datasources.total_datasources_by_type),
                backgroundColor: chartColors,
            },
        ],
    };

    const documentsFetchedData = {
        labels: Object.keys(analyticsData.datasources.total_documents_fetched_by_datasource_type),
        datasets: [
            {
                data: Object.values(analyticsData.datasources.total_documents_fetched_by_datasource_type),
                backgroundColor: chartColors,
            },
        ],
    };

    const chartOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                position: 'bottom' as const,
                labels: {
                    boxWidth: 12,
                    padding: 15,
                    font: {
                        size: 11
                    }
                },
            },
            tooltip: {
                callbacks: {
                    label: function (context: any) {
                        let label = context.label || '';
                        if (label) {
                            label += ': ';
                        }
                        if (context.parsed !== null) {
                            label += context.parsed;
                        }
                        return label;
                    }
                }
            }
        },
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6">
            {/* Embedding Providers Card */}
            <div className="bg-white p-6 rounded-lg shadow-lg hover:shadow-xl transition-shadow duration-300">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Embedding Providers</h2>
                <div className="space-y-2">
                    <p className="text-3xl font-bold text-gray-800">{analyticsData.embedding_providers.total_providers}</p>
                    <p className="text-sm text-gray-500">Total Providers</p>
                </div>
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <p className="font-semibold text-gray-700">Active Provider:</p>
                    <p className="text-sm text-gray-600">{analyticsData.embedding_providers.active_provider.name || 'None'}</p>
                    <p className="text-xs text-gray-500">Model: {analyticsData.embedding_providers.active_provider.model || 'N/A'}</p>
                </div>
            </div>

            {/* LLM Providers Card */}
            <div className="bg-white p-6 rounded-lg shadow-lg hover:shadow-xl transition-shadow duration-300">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">LLM Providers</h2>
                <div className="space-y-2">
                    <p className="text-3xl font-bold text-gray-800">{analyticsData.llm_providers.total_providers}</p>
                    <p className="text-sm text-gray-500">Total Providers</p>
                </div>
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <p className="font-semibold text-gray-700">Active Provider:</p>
                    <p className="text-sm text-gray-600">{analyticsData.llm_providers.active_provider.name || 'None'}</p>
                    <p className="text-xs text-gray-500">Model: {analyticsData.llm_providers.active_provider.model || 'N/A'}</p>
                </div>
            </div>

            {/* Datasources Card */}
            <div className="bg-white p-6 rounded-lg shadow-lg hover:shadow-xl transition-shadow duration-300">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Datasources</h2>
                <div className="space-y-2">
                    <p className="text-3xl font-bold text-gray-800">{analyticsData.datasources.total_datasources}</p>
                    <p className="text-sm text-gray-500">Total Datasources</p>
                </div>
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <p className="font-semibold text-gray-700">Configured Datasources:</p>
                    {analyticsData.datasources.configured_datasources && analyticsData.datasources.configured_datasources.length > 0 ? (
                        <ul className="mt-2 space-y-1">
                            {analyticsData.datasources.configured_datasources.map((ds, index) => (
                                <li key={index} className="text-sm">
                                    <span className="font-medium text-gray-700">{ds.name}</span>
                                    <span className="text-xs text-gray-500 ml-2">({ds.type})</span>
                                    <span
                                        className={`text-xs ml-2 ${ds.status === 'active' ? 'text-green-500' : 'text-red-500'}`}>
                                        {ds.status}
                                    </span>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="text-sm text-gray-500 mt-2">No configured datasources</p>
                    )}
                </div>
            </div>

            {/* Datasources Overview Charts */}
            <div
                className="bg-white p-6 rounded-lg shadow-lg hover:shadow-xl transition-shadow duration-300 col-span-1 md:col-span-2 lg:col-span-3">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Datasources Overview</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="h-[300px]">
                        <h3 className="text-lg font-semibold mb-2 text-gray-700">Datasources by Type</h3>
                        {Object.keys(analyticsData.datasources.total_datasources_by_type).length > 0 ? (
                            <Pie
                                data={datasourcesByTypeData}
                                options={chartOptions}
                            />
                        ) : (
                            <p className="text-center text-gray-500 mt-8">No data available</p>
                        )}
                    </div>
                    <div className="h-[300px]">
                        <h3 className="text-lg font-semibold mb-2 text-gray-700">Documents Fetched by Datasource
                            Type</h3>
                        {Object.keys(analyticsData.datasources.total_documents_fetched_by_datasource_type).length > 0 ? (
                            <Pie
                                data={documentsFetchedData}
                                options={chartOptions}
                            />
                        ) : (
                            <p className="text-center text-gray-500 mt-8">No data available</p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;