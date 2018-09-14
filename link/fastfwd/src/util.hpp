#ifndef __UTIL_HPP__
#define __UTIL_HPP__

#include <memory>
#include <algorithm>

#include <chrono>
#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>

//
// singleton
//

template <class T>
class Singular
{
public:
        template <class... Args>
        static T* Instance(Args... _args) {
                static auto getInstance = [_args...]() -> T& {
                        return GetInstance(_args...);
                };
                return GetOne(getInstance);
        }
protected:
        Singular() = delete;
        ~Singular() = delete;
        Singular(Singular&) = delete;
        Singular& operator=(const Singular&) = delete;
private:
        static T* GetOne(const std::function<T&()>& _getInstance) {
                static T& instance = _getInstance();
                return &instance;
        }
        template<class... Args>
        static T& GetInstance(Args... _args) {
                static T instance{std::forward<Args>(_args)...};
                return instance;
        }
};

//
// defer
//

typedef struct _Defer
{
        _Defer(const std::function<void()>& _deferredFunc): m_deferredFunc(_deferredFunc) {}
        ~_Defer() { m_deferredFunc(); }
private:
        const std::function<void()> m_deferredFunc;
        _Defer() = delete;
        _Defer(const _Defer&) = delete;
        _Defer(_Defer&&) = delete;
        _Defer &operator=(const _Defer&) = delete;
        _Defer &operator=(_Defer&&) = delete;
        void *operator new(std::size_t, ...) = delete;
        void operator delete(void*) = delete;
        void operator delete[](void*) = delete;
} Defer;

//
// queue
//

template <class T>
class SharedQueue
{
public:
        SharedQueue(size_t nItems);
        SharedQueue();
        T Pop();
        bool TryPop(T &item);
        bool PopWithTimeout(T& item, const std::chrono::milliseconds& timeout);
        void Push(const T& item);
        void Push(T&& item);
        bool TryPush(const T& item);
        bool ForcePush(const T& item);
        bool ForcePush(const T&& item);
        void Clear();

        // though I do not recommend using empty() or size() in concurrent environment
        // but it is natural to use it for debugging or verbose outputs purpose
        bool IsEmpty();
        size_t Size();

        // access to iterators
        size_t Foreach(const std::function<void(T&)> lambda);
        void FindIf(const std::function<bool(T&)> _lambda);
        void CriticalSection(const std::function<void(std::deque<T>&)> lambda);
private:
        std::deque<T> m_queue;
        std::mutex m_mutex;
        std::condition_variable m_consumerCondition;
        std::condition_variable m_producerCondition;
        size_t m_nMaxItems;
};

template <class T>
SharedQueue<T>::SharedQueue() :
        m_queue(std::deque<T>()),
        m_mutex(),
        m_consumerCondition(),
        m_producerCondition(),
        m_nMaxItems(0)
{}

template <class T>
SharedQueue<T>::SharedQueue(size_t _nItems):
        m_queue(std::deque<T>()),
        m_mutex(),
        m_consumerCondition(),
        m_producerCondition(),
        m_nMaxItems(_nItems)
{}

template <class T>
bool SharedQueue<T>::PopWithTimeout(T& _item, const std::chrono::milliseconds& _timeout)
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_queue.empty()) {
                auto ret = m_consumerCondition.wait_for(mutexLock, _timeout);
                if (ret == std::cv_status::timeout) {
                        return false; 
                }
        }
        if (!m_queue.empty()) {
                _item = m_queue.back();
                m_queue.pop_back();
                m_producerCondition.notify_one();
                return true;
        }
        return false;
}

template <class T>
T SharedQueue<T>::Pop()
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        while (m_queue.empty()) {
                m_consumerCondition.wait(mutexLock);
        }
        auto item = m_queue.back();
        m_queue.pop_back();
        m_producerCondition.notify_one();
        return item;
}

template <class T>
bool SharedQueue<T>::TryPop(T& _item)
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (! m_queue.empty()) {
                _item = m_queue.back();
                m_queue.pop_back();
                m_producerCondition.notify_one();
                return true;
        }
        return false;
}

template <class T>
void SharedQueue<T>::Push(const T& _item)
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_nMaxItems != 0 && m_queue.size() >= m_nMaxItems) {
                m_producerCondition.wait(mutexLock);
        }
        m_queue.push_front(_item);
        m_consumerCondition.notify_one();
}

template <class T>
bool SharedQueue<T>::ForcePush(const T& _item)
{
        bool bRet = true;
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_nMaxItems != 0 && m_queue.size() == m_nMaxItems) {
                m_queue.pop_back();
                bRet = false;
        }
        m_queue.push_front(_item);
        m_consumerCondition.notify_one();
        return bRet;
}

template <class T>
bool SharedQueue<T>::ForcePush(const T&& _item)
{
        bool bRet = true;
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_nMaxItems != 0 && m_queue.size() == m_nMaxItems) {
                m_queue.pop_back();
                bRet = false;
        }
        m_queue.push_front(std::move(_item));
        m_consumerCondition.notify_one();
        return bRet;
}

template <class T>
bool SharedQueue<T>::TryPush(const T& _item)
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_nMaxItems != 0 && m_queue.size() >= m_nMaxItems) {
                return false;
        }
        m_queue.push_front(_item);
        m_consumerCondition.notify_one();
        return true;
}

template <class T>
void SharedQueue<T>::Push(T&& _item)
{
        std::unique_lock<std::mutex> mutexLock(m_mutex);
        if (m_nMaxItems != 0 && m_queue.size() >= m_nMaxItems) {
                m_producerCondition.wait(mutexLock);
        }
        m_queue.push_front(std::move(_item));
        m_consumerCondition.notify_one();
}

template <class T>
bool SharedQueue<T>::IsEmpty()
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        return m_queue.empty();
}

template <class T>
void SharedQueue<T>::Clear()
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        m_queue.clear();
}

template <class T>
size_t SharedQueue<T>::Size()
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        return m_queue.size();
}

template <class T>
size_t SharedQueue<T>::Foreach(const std::function<void(T&)> _lambda)
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        for_each(m_queue.begin(), m_queue.end(), _lambda);
        return m_queue.size();
}

template <class T>
void SharedQueue<T>::FindIf(const std::function<bool(T&)> _lambda)
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        std::find_if(m_queue.begin(), m_queue.end(), _lambda);
}

template <class T>
void SharedQueue<T>::CriticalSection(const std::function<void(std::deque<T>&)> _lambda)
{
        std::lock_guard<std::mutex> mutexLock(m_mutex);
        _lambda(m_queue);
}

//
// map
//

template <typename K, typename V>
class SharedMap
{
public:
        SharedMap();
        SharedMap(const SharedMap&) = delete;
        void operator = (const SharedMap&) = delete;

        bool IsEmpty()const;
        size_t Size()const;

        bool Insert(const K&, const V&);
        bool Insert(const K&, V&&);
        void FindIf(const std::function<bool(const K&, V&)>);
        bool Find(const K&, const std::function<void(V&)>);
        bool Erase(const K&);
        bool Erase(const K&, V&);

        size_t Foreach(const std::function<void(const K&, V&)>);
        void CriticalSection(const std::function<void(std::unordered_map<K, V>&)>);

        // TODO
        //V operator [] (const K&);
        //V& operator [] (const K&);

private:
        std::unordered_map<K, V> map_;
        mutable std::mutex mutex_;
};

template <typename K, typename V>
SharedMap<K, V>::SharedMap()
        : map_(std::unordered_map<K, V>())
        , mutex_()
{}

template <typename K, typename V>
bool SharedMap<K, V>::Insert(const K& _key, const V& _val)
{
        std::unique_lock<std::mutex> lock(mutex_);
        auto p = map_.insert(std::pair<const K&, const V&>(_key, _val));
        return p.second;
}

template <typename K, typename V>
bool SharedMap<K, V>::Insert(const K& _key, V&& _val)
{
        std::unique_lock<std::mutex> lock(mutex_);
        auto p = map_.insert(std::pair<const K&, V&&>(_key, std::move(_val)));
        return p.second;
}

template <typename K, typename V>
bool SharedMap<K, V>::Erase(const K& _key)
{
        std::unique_lock<std::mutex> lock(mutex_);
        auto it = map_.find(_key);
        if (it != map_.end()) {
                map_.erase(it);
                return true;
        }
        return false;
}

template <typename K, typename V>
bool SharedMap<K, V>::Erase(const K& _key, V& _val)
{
        std::unique_lock<std::mutex> lock(mutex_);
        auto it = map_.find(_key);
        if (it != map_.end()) {
                _val = std::move(it->second);
                map_.erase(it);
                return true;
        }
        return false;
}

template <typename K, typename V>
size_t SharedMap<K, V>::Foreach(const std::function<void(const K&, V&)> _lambda)
{
        std::lock_guard<std::mutex> lock(mutex_);
        std::for_each(map_.begin(), map_.end(), [_lambda](std::pair<const K, V>& _p){
                        _lambda(_p.first, _p.second);
        });
        return map_.size();
}

template <typename K, typename V>
void SharedMap<K, V>::CriticalSection(const std::function<void(std::unordered_map<K, V>&)> _lambda)
{
        std::lock_guard<std::mutex> lock(mutex_);
        _lambda(map_);
}

template <typename K, typename V>
void SharedMap<K, V>::FindIf(const std::function<bool(const K&, V&)> _lambda)
{
        std::lock_guard<std::mutex> lock(mutex_);
        std::find_if(map_.begin(), map_.end(), [_lambda](std::pair<const K, V>& _p)->bool{
                return _lambda(_p.first, _p.second);
        });
}

template <typename K, typename V>
bool SharedMap<K, V>::Find(const K& _key, const std::function<void(V&)> _lambda)
{
        std::lock_guard<std::mutex> lock(mutex_);
        auto it = map_.find(_key);
        if (it != map_.end()) {
                _lambda(it->second);
                return true;
        }
        return false;
}

template <typename K, typename V>
bool SharedMap<K, V>::IsEmpty()const
{
        std::lock_guard<std::mutex> lock(mutex_);
        return map_.empty();
}

template <typename K, typename V>
size_t SharedMap<K, V>::Size()const
{
        std::lock_guard<std::mutex> lock(mutex_);
        return map_.size();
}

#endif
