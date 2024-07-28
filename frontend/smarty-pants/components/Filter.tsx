// File: /components/Filter.tsx

import React, {useState} from 'react';

interface FilterProps {
    onFilterApply: (status: string, limit: number) => void;
}

const Filter: React.FC<FilterProps> = ({onFilterApply}) => {
    const [status, setStatus] = useState<string>('');
    const [limit, setLimit] = useState<number>(10);

    const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        setStatus(e.target.value);
    };

    const handleLimitChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setLimit(Number(e.target.value));
    };

    const handleApplyFilter = () => {
        onFilterApply(status, limit);
    };

    return (
        <div className="flex flex-col md:flex-row justify-between items-center mb-6 space-y-4 md:space-y-0">
            <div className="flex flex-col md:flex-row space-y-4 md:space-y-0 md:space-x-4 w-full md:w-auto">
                <select
                    id="status"
                    value={status}
                    onChange={handleStatusChange}
                    className="block w-full md:w-auto rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                >
                    <option value="">All Status</option>
                    <option value="pending">Pending</option>
                    <option value="ready_to_search">Ready to Search</option>
                    <option value="error_processing">Error Processing</option>
                </select>
                <input
                    type="number"
                    id="limit"
                    value={limit}
                    onChange={handleLimitChange}
                    placeholder="Limit"
                    className="block w-full md:w-auto rounded-md border-0 py-1.5 pl-3 pr-3 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
            </div>
            <button
                onClick={handleApplyFilter}
                className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                Apply Filters
            </button>
        </div>
    );
};

export default Filter;