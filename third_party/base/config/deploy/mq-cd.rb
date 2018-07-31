role :app, "192.168.20.110" # sandbox
role :db,  "192.168.20.110", :primary => true # This is where Rails migrations will run
role :web, "192.168.20.110"                          # Your HTTP server, Apache/etc
set :user, 'qboxserver'
set :repository,  "http://ci.qbox.me/jenkins/job/#{application}-production"
set :deploy_to do "/home/#{user}/websites/mq" end
set :jenkins_artifact_file, 'dist/qboxmq'
set :conf_name, 'qboxmq'


namespace :deploy do
  task :more_more_setup do
    run "mkdir -p #{shared_path}/run/mq"
  end
end

after "deploy:more_setup", "deploy:more_more_setup"

namespace :deploy do
  task :restart, :roles => :app, :max_hosts => 1 do
    run 'supervisorctl restart mq'
  end
end


