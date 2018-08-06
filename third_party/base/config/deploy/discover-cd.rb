role :app, "192.168.20.103" # sandbox
role :db,  "192.168.20.103", :primary => true # This is where Rails migrations will run
role :web, "192.168.20.103" # Your HTTP server, Apache/etc
set :user, 'qboxserver'
set :repository,  "https://ci.qiniu.io/jenkins/job/#{application}-production"
set :deploy_to do "/home/#{user}/websites/discover" end
set :jenkins_artifact_file, 'dist/qboxdiscover'
set :conf_name, 'qboxdiscover'


namespace :deploy do
  task :more_more_setup do
    run "mkdir -p #{shared_path}/run/discover"
    run "mkdir -p #{shared_path}/run/auditlog"
  end
end

after "deploy:more_setup", "deploy:more_more_setup"

namespace :deploy do
  task :restart, :roles => :app, :max_hosts => 1 do
    run 'supervisorctl restart discover'
  end
end


