import SearchIcon from '@mui/icons-material/Search';
import React from 'react'
import InputField from '../../components/FormFields/InputField'
import PageLayout from '../../layouts/PageLayout'
import UserCard from './components/UserCard';

const ExploreUI: React.FC<any> = ({
    query,
    handleQuery,
    usersList
}) => {

    return (
        <PageLayout pageTitle="Explore">
            <InputField
                value={query}
                setValue={handleQuery}
                placeholder="Search users by username or name"
                leftIcon={<SearchIcon style={{ color: 'gray' }} />} />

            <div className='grid grid-cols-1 gap-4 lg:grid-cols-3 md:grid-cols-2'>
                {[1, 2, 3, 4, 5].map((user: any) => (
                    <UserCard />
                ))}
            </div>
        </PageLayout>
    )
}

export default ExploreUI