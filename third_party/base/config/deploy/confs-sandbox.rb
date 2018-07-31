role :app, "192.168.1.200" # sandbox
role :db,  "192.168.1.200", :primary => true # This is where Rails migrations will run
role :web, "192.168.1.200"                          # Your HTTP server, Apache/etc
set :user, 'qboxserver'
set :repository,  "http://ci.qbox.me/jenkins/job/#{application}"
set :deploy_to do "/home/#{user}/websites/confs" end
set :jenkins_artifact_file, 'dist/qboxconfs'
set :conf_name, 'qboxconfs'


namespace :deploy do
  task :restart, :roles => :app, :max_hosts => 1 do
    run 'supervisorctl restart confs'
  end
end


